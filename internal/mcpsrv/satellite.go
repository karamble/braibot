// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mcpsrv

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/brmcp/server"
	"github.com/karamble/satfetch"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	kit "github.com/vctt94/bisonbotkit"
)

// Satellite products are priced by the effective ground resolution of the
// delivered image: the finer the zoom, the higher the tier. Sentinel-2 is
// floored at its native 10 m and always lands on the cheapest tier; the
// centimeter-class ortho sources climb with zoom.
const satPriceNote = "Priced by effective ground detail (size_km*1000/px, floored by the " +
	"source's native resolution): 5 m/px or coarser $0.005, 1-5 $0.01, 0.5-1 $0.02, " +
	"0.25-0.5 $0.03, finer $0.05."

// sentinelGSD is Sentinel-2's native visual resolution in meters per pixel.
const sentinelGSD = 10

// satEmbedMaxBytes is the inline-embed budget; larger products (and all
// GeoTIFFs) go out as a file transfer instead.
const satEmbedMaxBytes = 900 << 10

func satTierUSD(mpp float64) float64 {
	switch {
	case mpp >= 5:
		return 0.005
	case mpp >= 1:
		return 0.01
	case mpp >= 0.5:
		return 0.02
	case mpp >= 0.25:
		return 0.03
	default:
		return 0.05
	}
}

// satEffectiveMPP is the ground resolution a request asks for, floored by
// what the source can natively deliver.
func satEffectiveMPP(sizeKM float64, px int, nativeGSD float64) float64 {
	if sizeKM <= 0 || px <= 0 {
		return nativeGSD
	}
	mpp := sizeKM * 1000 / float64(px)
	if mpp < nativeGSD {
		mpp = nativeGSD
	}
	return mpp
}

func orDefaultF(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

func orDefaultI(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

type satScenesIn struct {
	Lat      float64 `json:"lat" jsonschema:"latitude, -90..90"`
	Lon      float64 `json:"lon" jsonschema:"longitude, -180..180"`
	MaxCloud float64 `json:"max_cloud,omitempty" jsonschema:"max cloud cover percent, 0..100 (default 100)"`
	Days     int     `json:"days,omitempty" jsonschema:"lookback window in days, 1..365 (default 90)"`
	Limit    int     `json:"limit,omitempty" jsonschema:"max scenes, 1..50 (default 20)"`
}

type satImageIn struct {
	Lat      float64 `json:"lat" jsonschema:"latitude, -90..90"`
	Lon      float64 `json:"lon" jsonschema:"longitude, -180..180"`
	SizeKM   float64 `json:"size_km,omitempty" jsonschema:"square window edge in km, 0.5..50 (default 5)"`
	MaxCloud float64 `json:"max_cloud,omitempty" jsonschema:"max cloud cover percent, 0..100 (default 20)"`
	Days     int     `json:"days,omitempty" jsonschema:"lookback window in days, 1..365 (default 30)"`
	SceneID  string  `json:"scene_id,omitempty" jsonschema:"pin an exact scene id from satellite_scenes, skips the search"`
	Format   string  `json:"format,omitempty" jsonschema:"jpeg (default), png, or gtiff (georeferenced, delivered as a file transfer)"`
	MaxPx    int     `json:"max_px,omitempty" jsonschema:"bound output pixels per side, 64..4096 (default 1024)"`
}

type satNDVIIn struct {
	Lat      float64 `json:"lat" jsonschema:"latitude, -90..90"`
	Lon      float64 `json:"lon" jsonschema:"longitude, -180..180"`
	SizeKM   float64 `json:"size_km,omitempty" jsonschema:"square window edge in km, 0.5..50 (default 5)"`
	MaxCloud float64 `json:"max_cloud,omitempty" jsonschema:"max cloud cover percent, 0..100 (default 20)"`
	Days     int     `json:"days,omitempty" jsonschema:"lookback window in days, 1..365 (default 30)"`
	SceneID  string  `json:"scene_id,omitempty" jsonschema:"pin an exact scene id from satellite_scenes, skips the search"`
	Format   string  `json:"format,omitempty" jsonschema:"png (default, color ramp) or gtiff (Float32 values, delivered as a file transfer)"`
	MaxPx    int     `json:"max_px,omitempty" jsonschema:"bound output pixels per side, 64..4096 (default 1024)"`
}

type satOrthoIn struct {
	Lat    float64 `json:"lat" jsonschema:"latitude, -90..90"`
	Lon    float64 `json:"lon" jsonschema:"longitude, -180..180"`
	Source string  `json:"source" jsonschema:"ortho source name from satellite_sources, e.g. pl, nl, ch, us"`
	SizeKM float64 `json:"size_km,omitempty" jsonschema:"square window edge in km, 0.1..10 (default 1)"`
	Px     int     `json:"px,omitempty" jsonschema:"output width and height in pixels, 64..4096 (default 1024)"`
	Format string  `json:"format,omitempty" jsonschema:"jpeg (default) or png"`
}

// AttachSatellite registers the satfetch-backed imagery tools on the
// harness. All upstream data sources are free and keyless.
func AttachSatellite(h *server.Harness, svc *satfetch.Service, bot *kit.Bot) {
	sources := map[string]satfetch.SourceInfo{}
	for _, src := range svc.SourceCatalog() {
		sources[src.Name] = src
	}

	server.AddTool(h, &mcp.Tool{
		Name: "satellite_sources",
		Description: "List the centimeter-class aerial orthophoto sources for satellite_ortho: " +
			"name, type, native resolution in m/px, attribution. Coverage is per country. " +
			"Global Sentinel-2 imagery (satellite_image, 10 m) needs no source.",
	}, 0, func(_ context.Context, _ string, _ struct{}) (any, error) {
		out := make([]map[string]any, 0, len(sources))
		for _, src := range svc.SourceCatalog() {
			out = append(out, map[string]any{
				"name":        src.Name,
				"type":        src.Type,
				"gsd":         src.GSD,
				"attribution": src.Attribution,
			})
		}
		return map[string]any{"sources": out}, nil
	})

	server.AddTool(h, &mcp.Tool{
		Name: "satellite_scenes",
		Description: "List recent Sentinel-2 scenes covering a point: id, acquisition time, cloud " +
			"cover percent. Use a scene id with satellite_image or satellite_ndvi to pin an exact " +
			"acquisition.",
	}, 0, func(ctx context.Context, _ string, in satScenesIn) (any, error) {
		scenes, err := svc.Scenes(ctx, satfetch.ScenesRequest{
			Lat: in.Lat, Lon: in.Lon, MaxCloud: in.MaxCloud, Days: in.Days, Limit: in.Limit,
		})
		if err != nil {
			return nil, err
		}
		out := make([]map[string]any, 0, len(scenes))
		for _, sc := range scenes {
			out = append(out, map[string]any{
				"id":          sc.ID,
				"datetime":    sc.Datetime.UTC().Format(time.RFC3339),
				"cloud_cover": sc.CloudCover,
			})
		}
		return map[string]any{"scenes": out}, nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name: "satellite_image",
		Description: "True-color satellite image of any point on Earth (Sentinel-2, 10 m native " +
			"resolution, ~5 day revisit). The image is delivered to your Bison Relay DM with a " +
			"description; format gtiff arrives as a georeferenced file transfer. " + satPriceNote,
	}, func(_ context.Context, _ string, in satImageIn) (int64, error) {
		mpp := satEffectiveMPP(orDefaultF(in.SizeKM, 5), orDefaultI(in.MaxPx, 1024), sentinelGSD)
		return usdToAtoms(satTierUSD(mpp))
	}, func(ctx context.Context, peer string, in satImageIn) (any, error) {
		sizeKM := orDefaultF(in.SizeKM, 5)
		format := satfetch.Format(strings.ToLower(in.Format))
		if format == "" {
			format = satfetch.FormatJPEG
		}
		res, err := svc.Image(ctx, satfetch.ImageRequest{
			Lat: in.Lat, Lon: in.Lon, SizeKM: in.SizeKM, MaxCloud: in.MaxCloud,
			Days: in.Days, SceneID: in.SceneID,
			Format: format,
			MaxPx:  orDefaultI(in.MaxPx, 1024),
		})
		if err != nil {
			return nil, err
		}
		return satDeliver(ctx, bot, peer, res, "image",
			"Satellite image (Sentinel-2)", in.Lat, in.Lon, sizeKM, "")
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name: "satellite_ndvi",
		Description: "NDVI vegetation-index render of any point on Earth (Sentinel-2 red+nir " +
			"bands): browns and beiges are bare ground, greens are vegetation. Delivered to your " +
			"Bison Relay DM; format gtiff arrives as a georeferenced Float32 file. " + satPriceNote,
	}, func(_ context.Context, _ string, in satNDVIIn) (int64, error) {
		mpp := satEffectiveMPP(orDefaultF(in.SizeKM, 5), orDefaultI(in.MaxPx, 1024), sentinelGSD)
		return usdToAtoms(satTierUSD(mpp))
	}, func(ctx context.Context, peer string, in satNDVIIn) (any, error) {
		sizeKM := orDefaultF(in.SizeKM, 5)
		format := satfetch.Format(strings.ToLower(in.Format))
		if format == "" {
			format = satfetch.FormatPNG
		}
		res, err := svc.NDVI(ctx, satfetch.ImageRequest{
			Lat: in.Lat, Lon: in.Lon, SizeKM: in.SizeKM, MaxCloud: in.MaxCloud,
			Days: in.Days, SceneID: in.SceneID,
			Format: format,
			MaxPx:  orDefaultI(in.MaxPx, 1024),
		})
		if err != nil {
			return nil, err
		}
		return satDeliver(ctx, bot, peer, res, "ndvi",
			"NDVI vegetation index (Sentinel-2)", in.Lat, in.Lon, sizeKM, "")
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name: "satellite_ortho",
		Description: "Centimeter-class aerial orthophoto from a national open-data source (see " +
			"satellite_sources for names and coverage; orthophotos are flown on multi-year " +
			"cycles). Delivered to your Bison Relay DM with a description. " + satPriceNote,
	}, func(_ context.Context, _ string, in satOrthoIn) (int64, error) {
		var native float64
		if src, ok := sources[in.Source]; ok {
			native = src.GSD
		}
		mpp := satEffectiveMPP(orDefaultF(in.SizeKM, 1), orDefaultI(in.Px, 1024), native)
		return usdToAtoms(satTierUSD(mpp))
	}, func(ctx context.Context, peer string, in satOrthoIn) (any, error) {
		sizeKM := orDefaultF(in.SizeKM, 1)
		res, err := svc.Ortho(ctx, satfetch.OrthoRequest{
			Lat: in.Lat, Lon: in.Lon, Source: in.Source, SizeKM: in.SizeKM,
			Px:     orDefaultI(in.Px, 1024),
			Format: satfetch.Format(strings.ToLower(in.Format)),
		})
		if err != nil {
			return nil, err
		}
		attribution := ""
		if src, ok := sources[in.Source]; ok {
			attribution = src.Attribution
		}
		return satDeliver(ctx, bot, peer, res, "ortho",
			fmt.Sprintf("Orthophoto (%s)", res.Source), in.Lat, in.Lon, sizeKM, attribution)
	})
}

var satExtensions = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/tiff": "tif",
}

// satDeliver sends the finished product to the caller's DM - inline embed
// with a descriptive caption when it fits, file transfer otherwise - and
// builds the delivery confirmation (no URLs).
func satDeliver(ctx context.Context, bot *kit.Bot, peer string, res *satfetch.Result,
	kind, header string, lat, lon, sizeKM float64, attribution string) (map[string]any, error) {

	data, err := os.ReadFile(res.Path)
	if err != nil {
		return nil, fmt.Errorf("read product: %w", err)
	}

	mpp := 0.0
	if res.Width > 0 {
		mpp = sizeKM * 1000 / float64(res.Width)
	}
	lines := []string{
		header,
		fmt.Sprintf("Center: %.5f, %.5f - %.1f x %.1f km", lat, lon, sizeKM, sizeKM),
	}
	model := "sentinel-2"
	extra := map[string]any{
		"width":   res.Width,
		"height":  res.Height,
		"size_km": sizeKM,
	}
	if res.Scene.ID != "" {
		lines = append(lines, fmt.Sprintf("Scene %s, acquired %s, cloud cover %.1f%%",
			res.Scene.ID, res.Scene.Datetime.UTC().Format(time.RFC3339), res.Scene.CloudCover))
		extra["scene_id"] = res.Scene.ID
		extra["datetime"] = res.Scene.Datetime.UTC().Format(time.RFC3339)
		extra["cloud_cover"] = res.Scene.CloudCover
	}
	if res.Source != "" {
		model = res.Source
		detail := fmt.Sprintf("Native resolution %.2f m/px", res.SourceGSD)
		if attribution != "" {
			detail = fmt.Sprintf("%s, native %.2f m/px", attribution, res.SourceGSD)
		}
		lines = append(lines, detail)
		extra["source"] = res.Source
		extra["gsd"] = res.SourceGSD
	}
	if res.Width > 0 {
		lines = append(lines, fmt.Sprintf("Output: %dx%d px (~%.2f m/px)", res.Width, res.Height, mpp))
	}
	if atoms := server.ChargedAtoms(ctx); atoms > 0 {
		charged := dcrutil.Amount(atoms)
		lines = append(lines, fmt.Sprintf("Charged: %s", charged))
		extra["charged_dcr"] = charged.ToCoin()
	}
	caption := strings.Join(lines, "\n")

	if res.ContentType == "image/tiff" || len(data) > satEmbedMaxBytes {
		ext := satExtensions[res.ContentType]
		name := fmt.Sprintf("satellite-%s-%s.%s", kind, time.Now().UTC().Format("20060102-150405"), ext)
		if res.Scene.ID != "" {
			name = fmt.Sprintf("satellite-%s-%s.%s", kind, res.Scene.ID, ext)
		}
		dir, err := os.MkdirTemp("", "braibot-satellite-*")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(dir)
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return nil, err
		}
		if err := bot.SendPM(ctx, peer, caption); err != nil {
			return nil, fmt.Errorf("send caption: %w", err)
		}
		if err := bot.SendFile(ctx, peer, path); err != nil {
			return nil, fmt.Errorf("send file: %w", err)
		}
		extra["delivery"] = "file"
		extra["file"] = name
		return delivered(model, kind, extra), nil
	}

	alt := "satellite " + kind
	if res.Source != "" {
		alt = "orthophoto " + res.Source
	}
	msg := fmt.Sprintf("%s\n\n--embed[alt=%s,type=%s,data=%s]--",
		caption, alt, res.ContentType, base64.StdEncoding.EncodeToString(data))
	if err := bot.SendPM(ctx, peer, msg); err != nil {
		return nil, fmt.Errorf("send image: %w", err)
	}
	extra["delivery"] = "embed"
	return delivered(model, kind, extra), nil
}
