// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package mcpsrv exposes braibot's generation capabilities as MCP tools
// over Bison Relay via the brmcp harness. Tools reuse the existing service
// layer with its internal billing disabled: the harness quotes and debits
// braibot's own balance store per call (payment_required with an invoice or
// tip on shortfall), and results are delivered exclusively over the Bison
// Relay DM like every other braibot job - tool results carry no content
// URLs, only a delivery confirmation.
package mcpsrv

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/brmcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	kit "github.com/vctt94/bisonbotkit"

	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/image"
	"github.com/karamble/braibot/internal/speech"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/internal/video"
	"github.com/karamble/braibot/pkg/fal"
)

// braibot's balance store keeps milli-atoms (1e11 per DCR); the brmcp
// Billing contract speaks atoms.
const matomsPerAtom = 1000

// Billing adapts braibot's per-user balance database to server.Billing, so
// harness debits, refunds, tip credits, and invoice settlements all move
// the same balances the chat commands use.
type Billing struct {
	db    *database.DBManager
	debug bool
}

func NewBilling(db *database.DBManager, debug bool) *Billing {
	return &Billing{db: db, debug: debug}
}

func (b *Billing) Balance(uid string) int64 {
	matoms, err := b.db.GetBalance(uid)
	if err != nil {
		return 0
	}
	return matoms / matomsPerAtom
}

func (b *Billing) Credit(uid string, atoms int64) error {
	return b.db.UpdateBalance(uid, atoms*matomsPerAtom)
}

func (b *Billing) Debit(uid string, atoms int64) error {
	raw, err := hex.DecodeString(uid)
	if err != nil {
		return fmt.Errorf("bad uid: %w", err)
	}
	ok, err := b.db.CheckAndDeductBalance(raw, atoms*matomsPerAtom, b.debug)
	if err != nil && !ok {
		// The store reports a shortfall as an error string; the harness
		// needs the sentinel to build payment_required.
		if strings.Contains(err.Error(), "insufficient balance") {
			return server.ErrInsufficient
		}
		return err
	}
	if !ok {
		return server.ErrInsufficient
	}
	return nil
}

// usdToAtoms quotes a USD price in atoms at the current exchange rate
// (braibot's cached CoinGecko feed).
func usdToAtoms(usd float64) (int64, error) {
	dcr, err := utils.USDToDCR(usd)
	if err != nil {
		return 0, fmt.Errorf("exchange rate unavailable: %w", err)
	}
	amt, err := dcrutil.NewAmount(dcr)
	if err != nil {
		return 0, err
	}
	return int64(amt), nil
}

// resolveModel picks the explicit model or the caller's current default for
// the command type.
func resolveModel(name, commandType, uid string) (faladapter.AppModel, error) {
	if name != "" {
		m, ok := faladapter.GetModel(name, commandType)
		if !ok {
			return faladapter.AppModel{}, fmt.Errorf("unknown %s model %q", commandType, name)
		}
		return m, nil
	}
	m, ok := faladapter.GetCurrentModel(commandType, uid)
	if !ok {
		return faladapter.AppModel{}, fmt.Errorf("no default %s model", commandType)
	}
	return m, nil
}

// videoPriceUSD prices per-second models by the requested duration
// (defaulting to 5 seconds, the common fal default), flat otherwise.
func videoPriceUSD(m faladapter.AppModel, duration string) float64 {
	if !m.PerSecondPricing {
		return m.PriceUSD
	}
	d, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(duration), "s"))
	if err != nil || d <= 0 {
		d = 5
	}
	return m.PriceUSD * float64(d)
}

func genReq(commandType string, m faladapter.AppModel, peer string) (braibottypes.GenerationRequest, error) {
	var uid zkidentity.ShortID
	if err := uid.FromString(peer); err != nil {
		return braibottypes.GenerationRequest{}, fmt.Errorf("bad caller uid: %w", err)
	}
	// UserNick doubles as the delivery target; bisonbotkit's send APIs
	// resolve a 64-hex uid as well as a nick.
	return braibottypes.GenerationRequest{
		ModelName: m.Name,
		ModelType: commandType,
		UserNick:  peer,
		UserID:    uid,
		PriceUSD:  m.PriceUSD,
		IsPM:      true,
	}, nil
}

// externalBilling mirrors the harness debit into the delivery DMs: the
// services run with internal billing disabled, so without it they would
// report billing as disabled while the caller was in fact charged.
func externalBilling(ctx context.Context, db *database.DBManager, peer string, usd float64) *braibottypes.ExternalBilling {
	atoms := server.ChargedAtoms(ctx)
	if atoms <= 0 {
		return nil
	}
	eb := &braibottypes.ExternalBilling{
		ChargedDCR: dcrutil.Amount(atoms).ToCoin(),
		ChargedUSD: usd,
	}
	if raw, err := hex.DecodeString(peer); err == nil {
		if bal, err := db.GetUserBalance(raw); err == nil {
			eb.BalanceDCR = bal
		}
	}
	return eb
}

// delivered is the uniform tool result: content goes out over the Bison
// Relay DM only, never as URLs in the MCP result.
func delivered(model, kind string, extra map[string]any) map[string]any {
	out := map[string]any{
		"status":    "delivered",
		"model":     model,
		"kind":      kind,
		"deliverui": "bison relay dm",
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

type text2ImageIn struct {
	Prompt    string `json:"prompt" jsonschema:"the image prompt"`
	Model     string `json:"model,omitempty" jsonschema:"model name; omit for the default"`
	NumImages int    `json:"num_images,omitempty" jsonschema:"number of images (default 1)"`
}

type text2VideoIn struct {
	Prompt   string `json:"prompt" jsonschema:"the video prompt"`
	Model    string `json:"model,omitempty" jsonschema:"model name; omit for the default"`
	Duration string `json:"duration,omitempty" jsonschema:"video length in seconds, digits only (default 5)"`
}

type image2VideoIn struct {
	Prompt   string `json:"prompt,omitempty" jsonschema:"optional motion prompt"`
	ImageURL string `json:"image_url" jsonschema:"source image URL the model animates"`
	Model    string `json:"model,omitempty" jsonschema:"model name; omit for the default"`
	Duration string `json:"duration,omitempty" jsonschema:"video length in seconds, digits only (default 5)"`
}

type text2SpeechIn struct {
	Text    string `json:"text" jsonschema:"the text to speak"`
	Model   string `json:"model,omitempty" jsonschema:"model name; omit for the default"`
	VoiceID string `json:"voice_id,omitempty" jsonschema:"optional voice id"`
}

type listModelsIn struct {
	Type string `json:"type,omitempty" jsonschema:"filter: text2image, text2video, image2video or text2speech"`
}

// exposed are the command types served as MCP tools.
var exposed = []string{"text2image", "text2video", "image2video", "text2speech"}

// Attach registers braibot's MCP tools on the harness. Services are built
// with billing DISABLED: the harness already debited the quote, so the
// service only validates, generates, and delivers over the DM.
func Attach(h *server.Harness, falClient *fal.Client, db *database.DBManager, bot *kit.Bot, debug bool) {
	imageSvc := image.NewImageService(falClient, db, bot, debug, false)
	videoSvc := video.NewVideoService(falClient, db, bot, debug, false)
	speechSvc := speech.NewSpeechService(falClient, db, bot, debug, false)

	server.AddTool(h, &mcp.Tool{
		Name:        "list_models",
		Description: "List the available generation models with their USD prices per call (per second for per-second models).",
	}, 0, func(_ context.Context, _ string, in listModelsIn) (any, error) {
		types := exposed
		if in.Type != "" {
			types = []string{in.Type}
		}
		out := map[string]any{}
		for _, ct := range types {
			models, ok := faladapter.GetModels(ct)
			if !ok {
				continue
			}
			list := make([]map[string]any, 0, len(models))
			for _, m := range models {
				list = append(list, map[string]any{
					"name":        m.Name,
					"description": m.Description,
					"priceUsd":    m.PriceUSD,
					"perSecond":   m.PerSecondPricing,
				})
			}
			out[ct] = list
		}
		return out, nil
	})

	server.AddTool(h, &mcp.Tool{
		Name:        "balance",
		Description: "Show your braibot balance in DCR. Fund it by tipping the bot or paying a payment_required invoice.",
	}, 0, func(_ context.Context, peer string, _ struct{}) (any, error) {
		raw, err := hex.DecodeString(peer)
		if err != nil {
			return nil, err
		}
		dcr, err := db.GetUserBalance(raw)
		if err != nil {
			return nil, err
		}
		return map[string]any{"balanceDcr": dcr}, nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name:        "text2image",
		Description: "Generate image(s) from a prompt. The result is delivered to your Bison Relay DM.",
	}, func(_ context.Context, peer string, in text2ImageIn) (int64, error) {
		m, err := resolveModel(in.Model, "text2image", peer)
		if err != nil {
			return 0, err
		}
		n := in.NumImages
		if n <= 0 {
			n = 1
		}
		return usdToAtoms(m.PriceUSD * float64(n))
	}, func(ctx context.Context, peer string, in text2ImageIn) (any, error) {
		if strings.TrimSpace(in.Prompt) == "" {
			return nil, errors.New("prompt is required")
		}
		m, err := resolveModel(in.Model, "text2image", peer)
		if err != nil {
			return nil, err
		}
		base, err := genReq("text2image", m, peer)
		if err != nil {
			return nil, err
		}
		n := in.NumImages
		if n <= 0 {
			n = 1
		}
		base.ExternalBilling = externalBilling(ctx, db, peer, m.PriceUSD*float64(n))
		req := &image.ImageRequest{GenerationRequest: base, Prompt: in.Prompt, NumImages: in.NumImages}
		if _, err := imageSvc.GenerateImage(ctx, req); err != nil {
			return nil, err
		}
		return delivered(m.Name, "image", map[string]any{"images": n}), nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name:        "text2video",
		Description: "Generate a video from a prompt. Generation can take many minutes; the result is delivered to your Bison Relay DM.",
	}, func(_ context.Context, peer string, in text2VideoIn) (int64, error) {
		m, err := resolveModel(in.Model, "text2video", peer)
		if err != nil {
			return 0, err
		}
		return usdToAtoms(videoPriceUSD(m, in.Duration))
	}, func(ctx context.Context, peer string, in text2VideoIn) (any, error) {
		if strings.TrimSpace(in.Prompt) == "" {
			return nil, errors.New("prompt is required")
		}
		m, err := resolveModel(in.Model, "text2video", peer)
		if err != nil {
			return nil, err
		}
		base, err := genReq("text2video", m, peer)
		if err != nil {
			return nil, err
		}
		base.ExternalBilling = externalBilling(ctx, db, peer, videoPriceUSD(m, in.Duration))
		req := &video.VideoRequest{GenerationRequest: base, Prompt: in.Prompt, Duration: in.Duration}
		if _, err := videoSvc.GenerateVideo(ctx, req); err != nil {
			return nil, err
		}
		return delivered(m.Name, "video", nil), nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name:        "image2video",
		Description: "Animate a source image into a video. Generation can take many minutes; the result is delivered to your Bison Relay DM.",
	}, func(_ context.Context, peer string, in image2VideoIn) (int64, error) {
		m, err := resolveModel(in.Model, "image2video", peer)
		if err != nil {
			return 0, err
		}
		return usdToAtoms(videoPriceUSD(m, in.Duration))
	}, func(ctx context.Context, peer string, in image2VideoIn) (any, error) {
		if strings.TrimSpace(in.ImageURL) == "" {
			return nil, errors.New("image_url is required")
		}
		m, err := resolveModel(in.Model, "image2video", peer)
		if err != nil {
			return nil, err
		}
		base, err := genReq("image2video", m, peer)
		if err != nil {
			return nil, err
		}
		base.ExternalBilling = externalBilling(ctx, db, peer, videoPriceUSD(m, in.Duration))
		req := &video.VideoRequest{
			GenerationRequest: base,
			Prompt:            in.Prompt,
			ImageURL:          in.ImageURL,
			Duration:          in.Duration,
		}
		if _, err := videoSvc.GenerateVideo(ctx, req); err != nil {
			return nil, err
		}
		return delivered(m.Name, "video", nil), nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name:        "text2speech",
		Description: "Synthesize speech from text. The audio is delivered to your Bison Relay DM.",
	}, func(_ context.Context, peer string, in text2SpeechIn) (int64, error) {
		m, err := resolveModel(in.Model, "text2speech", peer)
		if err != nil {
			return 0, err
		}
		return usdToAtoms(m.PriceUSD)
	}, func(ctx context.Context, peer string, in text2SpeechIn) (any, error) {
		if strings.TrimSpace(in.Text) == "" {
			return nil, errors.New("text is required")
		}
		m, err := resolveModel(in.Model, "text2speech", peer)
		if err != nil {
			return nil, err
		}
		base, err := genReq("text2speech", m, peer)
		if err != nil {
			return nil, err
		}
		base.ExternalBilling = externalBilling(ctx, db, peer, m.PriceUSD)
		req := &speech.SpeechRequest{GenerationRequest: base, Text: in.Text, VoiceID: in.VoiceID}
		if _, err := speechSvc.GenerateSpeech(ctx, req); err != nil {
			return nil, err
		}
		return delivered(m.Name, "audio", nil), nil
	})
}
