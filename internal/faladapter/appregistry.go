package faladapter

import (
	"fmt"

	"github.com/karamble/braibot/pkg/fal"
)

// appModelMeta stores braibot-specific metadata for each model, keyed by model name.
type appModelMeta struct {
	PriceUSD         float64
	PerSecondPricing bool
	HelpDoc          string
}

var (
	// defaultModels stores the default model for each command type.
	defaultModels = map[string]string{
		"text2image":  "flux/schnell",
		"image2image": "ghiblify",
		"text2speech": "minimax-tts/text-to-speech",
		"image2video": "veo2",
		"text2video":  "kling-video-text",
		"audio2text":  "elevenlabs/speech-to-text/scribe-v2",
		"video2video": "kling-video-o3-edit",
		"multi2video": "seedance-2.0-reference",
	}

	// userModels stores per-user model selections: map[userID]map[modelType]modelName
	userModels = make(map[string]map[string]string)

	// modelMeta maps model name → braibot-specific metadata (pricing, help docs).
	modelMeta = map[string]appModelMeta{
		// ── text2image ──────────────────────────────────────────
		"fast-sdxl": {PriceUSD: 0.02, HelpDoc: "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate"},
		"hidream-i1-full": {PriceUSD: 0.10, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a futuristic city --negative_prompt blur --guidance_scale 7\n\nParameters:\n• prompt: Text description (required)\n• --negative_prompt: Things to avoid (optional, default: \"\")\n• --image_size: Output dimensions (default: square_hd). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 50)\n• --seed: Specific seed (optional)\n• --guidance_scale: Prompt adherence (default: 5.0)\n• --num_images: Number of images (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: jpeg, png (default: jpeg)"},
		"hidream-i1-dev": {PriceUSD: 0.06, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a futuristic city --negative_prompt blur\n\nParameters:\n• prompt: Text description (required)\n• --negative_prompt: Things to avoid (optional, default: \"\")\n• --image_size: Output dimensions (default: square_hd). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 28)\n• --seed: Specific seed (optional)\n• --num_images: Number of images (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: jpeg, png (default: jpeg)"},
		"hidream-i1-fast": {PriceUSD: 0.03, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a futuristic city --negative_prompt blur\n\nParameters:\n• prompt: Text description (required)\n• --negative_prompt: Things to avoid (optional, default: \"\")\n• --image_size: Output dimensions (default: square_hd). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 16)\n• --seed: Specific seed (optional)\n• --num_images: Number of images (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: jpeg, png (default: jpeg)"},
		"flux-pro/v1.1": {PriceUSD: 0.08, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a hyperrealistic cat --num_images 2 --image_size square\n\nParameters:\n• prompt: Text description of the image (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --seed: Specific seed for reproducibility (optional)\n• --num_images: Number of images to generate (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true). Use --enable_safety_checker=false to disable.\n• --safety_tolerance: Safety strictness (1-6, default: 2)\n• --output_format: Image format (jpeg, png. default: jpeg)"},
		"flux-pro/v1.1-ultra": {PriceUSD: 0.12, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image cinematic photo --aspect_ratio 9:16 --raw=true\n\nParameters:\n• prompt: Text description (required)\n• --seed: Specific seed (optional)\n• --num_images: Number of images (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --safety_tolerance: Safety strictness (1-6, default: 2)\n• --output_format: jpeg, png (default: jpeg)\n• --aspect_ratio: Output aspect ratio (default: 16:9). Options: 21:9, 16:9, 4:3, 3:2, 1:1, 2:3, 3:4, 9:16, 9:21\n• --raw: Generate less processed image (default: false)"},
		"flux/schnell": {PriceUSD: 0.02, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a hyperrealistic cat --num_images 2 --image_size square\n\nParameters:\n• prompt: Text description of the image (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 4)\n• --seed: Specific seed for reproducibility (optional)\n• --num_images: Number of images to generate (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true). Use --enable_safety_checker=false to disable."},
		"flux/dev": {PriceUSD: 0.025, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a futuristic city --num_images 2 --image_size square\n\nParameters:\n• prompt: Text description of the image (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 28)\n• --seed: Specific seed for reproducibility (optional)\n• --guidance_scale: Prompt adherence (default: 3.5)\n• --num_images: Number of images to generate (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: jpeg, png (default: jpeg)"},
		"flux-2": {PriceUSD: 0.04, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a hyperrealistic cat --num_images 2 --image_size square_hd\n\nParameters:\n• prompt: Text description of the image (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --guidance_scale: Prompt adherence (default: 2.5)\n• --num_inference_steps: Number of steps (default: 28)\n• --seed: Specific seed for reproducibility (optional)\n• --num_images: Number of images to generate (default: 1)\n• --acceleration: Speed level: none, regular, high (default: regular)\n• --enable_prompt_expansion: Expand prompt for better results (default: false)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: Image format (jpeg, png, webp. default: png)"},
		"flux-2-pro": {PriceUSD: 0.08, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a hyperrealistic cat --image_size square_hd\n\nParameters:\n• prompt: Text description of the image (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --seed: Specific seed for reproducibility (optional)\n• --enable_safety_checker: Enable safety filter (default: true). Use --enable_safety_checker=false to disable.\n• --safety_tolerance: Safety strictness (1-5, default: 2)\n• --output_format: Image format (jpeg, png. default: jpeg)\n\nNote: This model generates 1 image per request (num_images not supported)."},
		"stable-diffusion-v35-large": {PriceUSD: 0.065, HelpDoc: "Usage: !text2image [prompt] [--option value]...\nExample: !text2image a hyperrealistic portrait --negative_prompt blur --guidance_scale 5\n\nParameters:\n• prompt: Text description of the image (required)\n• --negative_prompt: Things to avoid (optional)\n• --image_size: Output dimensions (default: square_hd). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --num_inference_steps: Number of steps (default: 40)\n• --seed: Specific seed for reproducibility (optional)\n• --guidance_scale: Prompt adherence (default: 4.5)\n• --num_images: Number of images to generate (default: 1)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --prompt_expansion: Use prompt expansion (default: true)\n• --output_format: jpeg, png (default: jpeg)"},

		// ── image2image ─────────────────────────────────────────
		"ghiblify":    {PriceUSD: 0.02, HelpDoc: "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform"},
		"cartoonify":  {PriceUSD: 0.02, HelpDoc: "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform"},
		"flux-2/edit": {PriceUSD: 0.06, HelpDoc: "Usage: !image2image [image_url] [prompt]\nExample: !image2image https://example.com/photo.jpg Add sunglasses to the person\n\nParameters:\n• image_url: URL of the source image (required, max 4 images)\n• prompt: Description of the desired edit (required)\n• --image_size: Output dimensions (default: landscape_4_3). Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --guidance_scale: Prompt adherence (default: 2.5)\n• --num_inference_steps: Number of steps (default: 28)\n• --seed: Specific seed for reproducibility (optional)\n• --num_images: Number of images to generate (default: 1)\n• --acceleration: Speed level: none, regular, high (default: regular)\n• --enable_prompt_expansion: Expand prompt for better results (default: false)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --output_format: Image format (jpeg, png, webp. default: png)"},
		"flux-2-pro/edit": {PriceUSD: 0.09, HelpDoc: "Usage: !image2image [image_url] [prompt]\nExample: !image2image https://example.com/photo.jpg Place realistic flames emerging from the top of the coffee cup\n\nParameters:\n• image_url: URL of the source image (required)\n• prompt: Description of the desired edit (required)\n• --image_size: Output dimensions (default: auto). Options: auto, square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9\n• --seed: Specific seed for reproducibility (optional)\n• --enable_safety_checker: Enable safety filter (default: true)\n• --safety_tolerance: Safety strictness (1-5, default: 2)\n• --output_format: Image format (jpeg, png. default: jpeg)"},
		"star-vector": {PriceUSD: 1.0, HelpDoc: "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nTo use this model, first set it as the default model:\n!setmodel image2image star-vector\n\nParameters:\n• image_url: URL of the source image\n\nPricing:\n• Base price: $1.0 per image"},

		// ── text2video ──────────────────────────────────────────
		"kling-video-text":            {PriceUSD: 0.4, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.40 per video."},
		"minimax/video-01-director":   {PriceUSD: 0.8, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.80 per video."},
		"minimax/video-01":            {PriceUSD: 0.8, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.80 per video"},
		"minimax/hailuo-02":           {PriceUSD: 0.09, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [--duration 6|10] [--prompt_optimizer true|false]\n\n\U0001f4b0 **Price: $0.10 per video second**\nExample: A 10-second video will cost $1.00.\nTotal cost = price per second \u00d7 duration."},
		"hunyuan-video":               {PriceUSD: 0.50, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.50 per video\n\nParameters:\n• prompt: Text description (required)\n• --aspect_ratio: 16:9, 9:16, 4:3, 3:4, 1:1 (default: 16:9)\n• --resolution: 480p, 580p, 720p, 1080p (default: 720p)\n• --video_length: 5s, 10s (default: 5s)\n• --num_inference_steps: Number of steps (default: 50)\n• --seed: Specific seed (optional)\n• --enable_safety_checker: Enable safety filter (default: true)"},
		"kling-video-v25-text":        {PriceUSD: 0.32, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.32 per second\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (5 or 10, default: 5)\n• --aspect_ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)"},
		"kling-video-v3-text":         {PriceUSD: 0.30, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.30 per second**\nExample: A 5-second video will cost $1.50.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)"},
		"kling-video-v3-pro-text":     {PriceUSD: 0.39, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.39 per second**\nExample: A 5-second video will cost $1.95.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)"},
		"kling-video-o3-text":         {PriceUSD: 0.28, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.28 per second**\nExample: A 5-second video will cost $1.40.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --audio: Enable audio generation (default: true)"},
		"kling-video-o3-pro-text":     {PriceUSD: 0.33, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.33 per second**\nExample: A 5-second video will cost $1.65.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --audio: Enable audio generation (default: true)"},
		"seedance-2.0-text":           {PriceUSD: 0.35, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.35 per second**\nExample: A 5-second video will cost $1.75.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description of the desired video (required)\n• --duration: Video duration in seconds (4-15, default: 5)\n• --aspect: Aspect ratio (auto, 21:9, 16:9, 4:3, 1:1, 3:4, 9:16). Default: auto\n• --resolution: Video resolution (480p, 720p). Default: 720p\n• --audio: Enable audio generation (default: true)\n• --seed: Seed for reproducibility (optional)"},
		"grok-imagine-video-text":     {PriceUSD: 0.08, PerSecondPricing: true, HelpDoc: "Usage: !text2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.08 per video second**\nExample: A 6-second video will cost $0.48.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description (required, max 4096 chars)\n• --duration: Video duration in seconds (1-15, default: 6)\n• --aspect: Aspect ratio: 16:9, 4:3, 3:2, 1:1, 2:3, 3:4, 9:16 (default: 16:9)\n• --resolution: 480p, 720p (default: 720p)"},

		// ── image2video ─────────────────────────────────────────
		"veo2":                              {PriceUSD: 3.50, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --aspect 16:9 --duration 5\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (16:9, 9:16, 1:1)\n• --duration: Video duration (5, 6, 7, 8)\n\nPricing:\n• Base price: $3.50 for 5 seconds\n• Additional cost: $0.50 per second beyond 5 seconds"},
		"kling-video-image":                 {PriceUSD: 2.0, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 10 --aspect 16:9\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --duration: Video duration in seconds (default: 5, min: 5)\n• --aspect: Aspect ratio (default: 16:9)\n• --negative-prompt: Text describing what to avoid (default: blur, distort, and low quality)\n• --cfg-scale: Configuration scale (default: 0.5)\n\nPricing:\n• Base price: $2.0 for 5 seconds\n• Additional cost: $0.4 per second beyond 5 seconds"},
		"minimax/video-01-subject-reference": {PriceUSD: 0.8, HelpDoc: "Usage: !image2video [subject_reference_image_url] [prompt] [options]\nExample: !image2video https://example.com/subject.jpg a person walking --prompt-optimizer false\n\nParameters:\n• subject_reference_image_url: URL of the image to use for consistent subject appearance.\n• prompt: Description of the desired video animation.\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)"},
		"minimax/video-01-live":              {PriceUSD: 0.8, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.png A character waving --prompt-optimizer true\n\nInfo: This model is specialized in bringing 2D illustrations to life.\n\nParameters:\n• image_url: URL of the image to animate.\n• prompt: Description of the desired video animation.\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)"},
		"veo3":                               {PriceUSD: 0.45, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 8s --resolution 1080p --audio\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (auto, 16:9, 9:16). Default: 16:9\n• --duration: Video duration (4s, 6s, 8s). Default: 8s\n• --resolution: Video resolution (720p, 1080p). Default: 720p\n• --audio: Enable audio generation. Default: true\n• --auto-fix: Auto-fix failed prompts. Default: false\n\nPricing:\n• $0.45 per second of video generated"},
		"veo31fast":                           {PriceUSD: 0.10, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 8s --resolution 1080p --audio\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (auto, 16:9, 9:16). Default: auto\n• --duration: Video duration (4s, 6s, 8s). Default: 8s\n• --resolution: Video resolution (720p, 1080p). Default: 720p\n• --audio: Enable audio generation. Default: true\n• --auto-fix: Auto-fix failed prompts. Default: false\n\nPricing:\n• $0.10 per second (no audio)\n• $0.15 per second (with audio)"},
		"kling-video-v25-image":               {PriceUSD: 0.32, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.32 per second\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired animation\n• --duration: Video duration in seconds (5 or 10, default: 5)\n• --aspect_ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)"},
		"luma-dream-machine":                  {PriceUSD: 0.40, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.40 per video\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired animation\n• --aspect_ratio: 16:9, 9:16, 4:3, 3:4, 21:9, 9:21, 1:1 (default: 16:9)\n• --loop: Create looping video (default: false)"},
		"ltx-video-13b":                       {PriceUSD: 0.30, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.30 per video\n\nParameters:\n• image_url: URL of the source image (for first/last frame)\n• prompt: Description of the desired animation\n• --num_frames: Number of frames (default: 97)\n• --frame_rate: Frame rate (default: 25)\n• --num_inference_steps: Number of steps (default: 30)\n• --guidance_scale: Prompt adherence (default: 3.0)\n• --negative_prompt: Things to avoid (optional)\n• --seed: Specific seed (optional)\n• --enable_safety_checker: Enable safety filter (default: true)"},
		"grok-imagine-video":                  {PriceUSD: 0.08, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 6 --aspect auto --resolution 720p\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --duration: Video duration in seconds (1-15, default: 6)\n• --aspect: Aspect ratio (auto, 16:9, 4:3, 3:2, 1:1, 2:3, 3:4, 9:16). Default: auto\n• --resolution: Video resolution (480p, 720p). Default: 720p\n\nPricing:\n• $0.08 per second of video generated"},
		"kling-video-v3-image":                {PriceUSD: 0.30, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.30 per second**\nExample: A 5-second video will cost $1.50.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• image_url: URL of the source image (required)\n• prompt: Description of the desired animation (optional)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)\n• --end_image: URL of end frame image (optional)"},
		"kling-video-v3-pro-image":            {PriceUSD: 0.39, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.39 per second**\nExample: A 5-second video will cost $1.95.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• image_url: URL of the source image (required)\n• prompt: Description of the desired animation (optional)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)\n• --end_image: URL of end frame image (optional)"},
		"seedance-2.0-image":                  {PriceUSD: 0.35, PerSecondPricing: true, HelpDoc: "Usage: !image2video [image_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.35 per second**\nExample: A 5-second video will cost $1.75.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• image_url: URL of the source image (required)\n• prompt: Description of the desired motion/action (required)\n• --duration: Video duration in seconds (4-15, default: 5)\n• --aspect: Aspect ratio (auto, 21:9, 16:9, 4:3, 1:1, 3:4, 9:16). Default: auto\n• --resolution: Video resolution (480p, 720p). Default: 720p\n• --audio: Enable audio generation (default: true)\n• --end_image: URL of end frame image (optional transition)\n• --seed: Seed for reproducibility (optional)"},

		// ── video2video ─────────────────────────────────────────
		"topaz-upscale-video":            {PriceUSD: 0.50, HelpDoc: "Usage: !video2video [video_url] [options]\n\n\U0001f4b0 **Price: $0.50 per video\n\nParameters:\n• video_url: URL of the video to upscale\n• --model: Upscaling model (default: auto)\n• --output_type: Output format mp4 or mov (default: mp4)"},
		"sync-lipsync-v2":                {PriceUSD: 0.10, PerSecondPricing: true, HelpDoc: "Usage: !video2video [video_url] [audio_url] [options]\n\n\U0001f4b0 **Price: $0.10 per second\n\nParameters:\n• video_url: URL of the video with face\n• audio_url: URL of the audio to sync\n• --model: wav2lip or wav2lip_gan (default: wav2lip)\n• --output_type: Output format mp4 or webm (default: mp4)"},
		"kling-video-v26-motion-control": {PriceUSD: 1.50, PerSecondPricing: true, HelpDoc: "Usage: !video2video [image_url] [video_url] [options]\n\n\U0001f4b0 **Price: $0.10 per second\n\nParameters:\n• image_url: Reference image URL (character/background source)\n• video_url: Reference video URL (motion source)\n• --prompt: Text description (optional)\n• --orientation: 'image' (max 10s) or 'video' (max 30s). Default: video\n• --keep-sound: Keep original audio (default: true)\n\nConstraints:\n• Character must occupy >5% of image with visible body"},
		"kling-video-o3-edit":            {PriceUSD: 0.30, PerSecondPricing: true, HelpDoc: "Usage: !video2video [video_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.30 per second**\nExample: A 5-second video will cost $1.50.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• video_url: URL of the source video (.mp4/.mov, 3-10s, 720-2160px)\n• prompt: Edit description (required, use @Image1-4 to reference images)\n• --keep_audio: Keep original audio (default: true)\n• --image1..--image4: Up to 4 reference image URLs\n• --duration: Duration for billing estimation (default: 5)"},
		"kling-video-o3-pro-edit":        {PriceUSD: 0.39, PerSecondPricing: true, HelpDoc: "Usage: !video2video [video_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.39 per second**\nExample: A 5-second video will cost $1.95.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• video_url: URL of the source video (.mp4/.mov, 3-10s, 720-2160px)\n• prompt: Edit description (required, use @Image1-4 to reference images)\n• --keep_audio: Keep original audio (default: true)\n• --image1..--image4: Up to 4 reference image URLs\n• --duration: Duration for billing estimation (default: 5)"},

		// ── multi2video ─────────────────────────────────────────
		"seedance-2.0-reference": {PriceUSD: 0.80, PerSecondPricing: true, HelpDoc: "Usage: !multi2video [prompt] [options]\n\n\U0001f4b0 **Price: $0.80 per second**\nExample: A 5-second video will cost $4.00.\nTotal cost = price per second \u00d7 duration.\n\nParameters:\n• prompt: Text description of the desired video (required)\n• --image1..--image9: Reference image URLs (up to 9, JPEG/PNG/WebP, max 30MB each)\n• --video1..--video3: Reference video URLs (up to 3, MP4/MOV, 2-15s combined duration, <50MB total, 480p-720p)\n• --audio1..--audio3: Reference audio URLs (up to 3, MP3/WAV, \u226415s combined, max 15MB each)\n• --duration: Output video duration in seconds (4-15, default: 5)\n• --aspect: Aspect ratio (auto, 21:9, 16:9, 4:3, 1:1, 3:4, 9:16). Default: auto\n• --resolution: Output video resolution (480p, 720p). Default: 720p\n• --audio: Enable generated audio output (default: true)\n• --seed: Seed for reproducibility (optional)\n\nConstraints:\n• At least one reference input (image, video, or audio) is required\n• Total reference files must not exceed 12\n• Reference audio requires at least one reference image or video"},

		// ── text2speech ─────────────────────────────────────────
		"minimax-tts/text-to-speech": {PriceUSD: 0.10, HelpDoc: "Usage: !text2speech [text] --voice_id [voice_id] [--option value]...\nExample: !text2speech Hello world --voice_id Wise_Woman --speed 0.8 --format flac\n\nParameters:\n• text: Text to convert to speech (required)\n• --voice_id: Voice ID to use (defaults to Wise_Woman if not specified). See list below.\n• --speed: Speech speed (0.5-2.0, default: 1.0)\n• --vol: Volume (0-10, default: 1.0)\n• --pitch: Voice pitch (-12 to 12, optional)\n• --emotion: happy, sad, angry, fearful, disgusted, surprised, neutral (optional)\n• --sample_rate: 8000, 16000, 22050, 24000, 32000, 44100 (default: 32000)\n• --bitrate: 32000, 64000, 128000, 256000 (default: 128000)\n• --format: mp3, pcm, flac (default: mp3)\n• --channel: 1 (mono), 2 (stereo) (default: 1)\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl"},
		"chatterbox-tts":             {PriceUSD: 0.05, HelpDoc: "Usage: !text2speech [text] [options]\n\n\U0001f4b0 **Price: $0.05 per 1000 characters\n\nParameters:\n• text: Text to convert to speech (required)\n• --audio_prompt_url: Reference audio URL for voice cloning (optional)\n• --exaggeration: Expression intensity 0-1 (default: 0.5)\n• --cfg_weight: Adherence to prompt 0-1 (default: 0.5)"},
		"elevenlabs-dialog":          {PriceUSD: 0.30, HelpDoc: "Usage: !text2speech [text] [options]\n\n\U0001f4b0 **Price: $0.30 per 1000 characters\n\nParameters:\n• text: Dialogue text with speaker labels (required)\n• --voice_id: Voice ID (default: Rachel)\n• --output_format: Audio format (default: mp3_22050_32)\n• --stability: Voice stability 0-1 (default: 0.5)\n• --similarity_boost: Voice similarity 0-1 (default: 0.75)"},
		"elevenlabs/tts/turbo-v2.5":  {PriceUSD: 0.05, HelpDoc: "Usage: !text2speech [text] [options]\n\n\U0001f4b0 **Price: $0.05 per 1000 characters\n\nParameters:\n• text: Text to convert to speech (required, max 5000 chars)\n• --voice: Voice name (default: Rachel)\n• --stability: Voice stability 0-1 (default: 0.5)\n• --similarity_boost: Voice similarity 0-1 (default: 0.75)\n• --style: Style exaggeration 0-1 (default: 0.0)\n• --speed: Speech speed 0.25-4.0 (default: 1.0)\n• --language_code: Language code (optional)\n\nAvailable Voices:\n• Aria, Roger, Sarah, Laura, Charlie, George, Callum\n• River, Liam, Charlotte, Alice, Matilda, Will, Jessica\n• Eric, Chris, Brian, Daniel, Lily, Bill"},

		// ── audio2text ──────────────────────────────────────────
		"elevenlabs/speech-to-text/scribe-v2": {PriceUSD: 0.008, PerSecondPricing: true, HelpDoc: "Usage: Transcribe audio to text with word-level timestamps\n\nPrice: $0.008 per minute of audio (plus 30% for diarization)\n\nParameters:\n- audio_url: URL to audio file (required)\n- task: transcribe (default) or translate\n- language: ISO 639-1 code (auto-detected if not specified)\n- chunk_level: segment (default) or word\n- diarize: Enable speaker diarization (default: true)\n- num_speakers: Number of speakers (optional, 1-50)\n\nSupported formats: mp3, wav, m4a, ogg, flac, webm"},

		// ── text2music ──────────────────────────────────────────
		"minimax-music-v2": {PriceUSD: 0.01, PerSecondPricing: true, HelpDoc: "Usage: !text2music [prompt] [options]\n\n\U0001f4b0 **Price: $0.01 per second of music\n\nParameters:\n• prompt: Description of the music (required)\n• --duration: Duration in seconds 1-300 (default: 60)\n• --reference_audio_url: URL of reference audio (optional)"},
		"stable-audio-25":  {PriceUSD: 0.02, PerSecondPricing: true, HelpDoc: "Usage: !text2music [prompt] [options]\n\n\U0001f4b0 **Price: $0.02 per second of audio\n\nParameters:\n• prompt: Description of the audio (required)\n• --duration: Duration in seconds 1-180 (default: 30)\n• --sample_rate: Sample rate (default: 44100)\n• --output_format: wav, mp3, ogg (default: wav)\n• --seed: Specific seed (optional)"},

		// ── audio2audio ─────────────────────────────────────────
		"elevenlabs-voice-changer": {PriceUSD: 0.50, PerSecondPricing: true, HelpDoc: "Usage: !audio2audio [audio_url] [options]\n\nPrice: $0.30 per minute\n\nParameters:\n- audio_url: URL of audio to transform (required)\n- --voice: Voice name (default: Rachel)\n- --remove_background_noise: Remove background noise (optional)\n- --seed: Random seed for reproducibility (optional)\n- --output_format: Output format (default: mp3_44100_128)\n\nAvailable Voices:\n- Aria, Roger, Sarah, Laura, Charlie, George, Callum\n- River, Liam, Charlotte, Alice, Matilda, Will, Jessica\n- Eric, Chris, Brian, Daniel, Lily, Bill, Rachel"},

		// ── video2audio ─────────────────────────────────────────
		"mmaudio-v2": {PriceUSD: 0.20, HelpDoc: "Usage: !video2audio [video_url] [prompt] [options]\n\n\U0001f4b0 **Price: $0.20 per video\n\nParameters:\n• video_url: URL of the source video\n• prompt: Description of the desired audio (optional)\n• --duration: Output duration in seconds (default: video duration)\n• --num_inference_steps: Number of steps (default: 25)\n• --seed: Specific seed (optional)"},
	}
)

// mergeAppModel combines a fal.Model with its braibot-specific metadata.
func mergeAppModel(m fal.Model) AppModel {
	meta, ok := modelMeta[m.Name]
	if !ok {
		return AppModel{Model: m}
	}
	return AppModel{
		Model:            m,
		PriceUSD:         meta.PriceUSD,
		PerSecondPricing: meta.PerSecondPricing,
		HelpDoc:          meta.HelpDoc,
	}
}

// GetModel returns an AppModel by name and type.
func GetModel(name, modelType string) (AppModel, bool) {
	m, ok := fal.GetModel(name, modelType)
	if !ok {
		return AppModel{}, false
	}
	return mergeAppModel(m), true
}

// GetModels returns all available AppModels for a command type.
func GetModels(commandType string) (map[string]AppModel, bool) {
	falModels, ok := fal.GetModels(commandType)
	if !ok {
		return nil, false
	}
	result := make(map[string]AppModel, len(falModels))
	for name, m := range falModels {
		result[name] = mergeAppModel(m)
	}
	return result, true
}

// GetCurrentModel returns the current model for a command type,
// checking per-user preferences first, then global defaults.
func GetCurrentModel(commandType string, userID string) (AppModel, bool) {
	var modelName string

	// Check user-specific model if userID is provided
	if userID != "" {
		if um, ok := userModels[userID]; ok {
			if name, ok := um[commandType]; ok {
				modelName = name
			}
		}
	}

	// Fall back to global default
	if modelName == "" {
		var ok bool
		modelName, ok = defaultModels[commandType]
		if !ok {
			return AppModel{}, false
		}
	}

	return GetModel(modelName, commandType)
}

// SetCurrentModel sets the current model for a command type.
func SetCurrentModel(commandType, modelName string, userID string) error {
	// Verify the model exists in the fal registry
	if _, ok := fal.GetModel(modelName, commandType); !ok {
		return fmt.Errorf("model not found: %s", modelName)
	}

	if userID != "" {
		if _, ok := userModels[userID]; !ok {
			userModels[userID] = make(map[string]string)
		}
		userModels[userID][commandType] = modelName
	} else {
		defaultModels[commandType] = modelName
	}
	return nil
}
