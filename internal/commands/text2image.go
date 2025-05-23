package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/image"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Text2ImageCommand returns the text2image command
// It now requires an ImageService instance.
func Text2ImageCommand(bot *kit.Bot, cfg *botconfig.BotConfig, imageService *image.ImageService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2image", "") // Empty string for global default
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "text2image",
			Description: "Generate an image from text using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !text2image [prompt]", model.Description)

	return braibottypes.Command{
		Name:        "text2image",
		Description: description,
		Category:    "AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Create a message sender using the adapter
			msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

			if len(args) < 1 {
				// Get the current model
				var userIDStr string
				if msgCtx.IsPM {
					var uid zkidentity.ShortID
					uid.FromBytes(msgCtx.Uid)
					userIDStr = uid.String()
				}
				model, exists := faladapter.GetCurrentModel("text2image", userIDStr)
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default text2image model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("text2image", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !text2image [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			// Parse arguments and prompt
			prompt, parsedReq, err := parseTextImageArgs(args)
			if err != nil {
				return msgSender.SendMessage(ctx, msgCtx, err.Error())
			}

			// Model config is needed for PriceUSD
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}
			model, exists := faladapter.GetCurrentModel("text2image", userIDStr)
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for text2image"))
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "text2image", msgCtx.IsPM, msgCtx.GC)

			// Create image request
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &image.ImageRequest{
				Prompt:              prompt,
				ModelType:           "text2image",
				ModelName:           model.Name,
				Progress:            progress,
				UserNick:            msgCtx.Nick,
				UserID:              userID,
				PriceUSD:            model.PriceUSD,
				NumImages:           parsedReq.NumImages,
				ImageSize:           parsedReq.ImageSize,
				Seed:                parsedReq.Seed,
				NumInferenceSteps:   parsedReq.NumInferenceSteps,
				EnableSafetyChecker: parsedReq.EnableSafetyChecker,
				SafetyTolerance:     parsedReq.SafetyTolerance,
				OutputFormat:        parsedReq.OutputFormat,
				NegativePrompt:      parsedReq.NegativePrompt,
				GuidanceScale:       parsedReq.GuidanceScale,
				AspectRatio:         parsedReq.AspectRatio,
				Raw:                 parsedReq.Raw,
				IsPM:                msgCtx.IsPM,
				GC:                  msgCtx.GC,
			}

			// Generate image using the service
			result, err := imageService.GenerateImage(ctx, req)

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, msgCtx, "text2image", result, err); handleErr != nil {
				return handleErr // Propagate error if not handled by the utility function
			}

			// If we reach here, the operation was successful and errors were handled
			return nil
		}),
	}
}

// parseTextImageArgs parses the command arguments for text2image, separating the prompt
// from known options.
// It returns the prompt string, a partially populated ImageRequest struct containing
// parsed options, and an error if parsing fails.
func parseTextImageArgs(args []string) (string, *image.ImageRequest, error) {
	var promptParts []string
	parsedReq := &image.ImageRequest{
		NumImages: 1, // Default
		// Initialize pointers/zero values for optional fields
		ImageSize:           "",
		Seed:                nil,
		NumInferenceSteps:   nil,
		EnableSafetyChecker: nil,
		SafetyTolerance:     "",
		OutputFormat:        "",
		NegativePrompt:      "",
		GuidanceScale:       nil,
		AspectRatio:         "",
		Raw:                 nil,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		argLower := strings.ToLower(arg)

		// Handle boolean flags like --flag=value
		var flagValue string
		if strings.Contains(argLower, "=") {
			parts := strings.SplitN(argLower, "=", 2)
			argLower = parts[0]
			if len(parts) > 1 {
				flagValue = parts[1]
			}
		}

		switch argLower {
		case "--num_images":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil || val <= 0 {
					return "", nil, fmt.Errorf("invalid value for --num_images: '%s'. Must be a positive integer", args[i+1])
				}
				parsedReq.NumImages = val
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --num_images argument")
			}
		case "--image_size":
			if i+1 < len(args) {
				parsedReq.ImageSize = args[i+1]
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --image_size argument")
			}
		case "--seed":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil {
					return "", nil, fmt.Errorf("invalid value for --seed: '%s'. Must be an integer", args[i+1])
				}
				parsedReq.Seed = &val
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --seed argument")
			}
		case "--num_inference_steps":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil || val <= 0 {
					return "", nil, fmt.Errorf("invalid value for --num_inference_steps: '%s'. Must be a positive integer", args[i+1])
				}
				parsedReq.NumInferenceSteps = &val
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --num_inference_steps argument")
			}
		case "--enable_safety_checker":
			var val bool
			var err error
			if flagValue != "" { // Handle --flag=value
				val, err = strconv.ParseBool(flagValue)
				if err != nil {
					return "", nil, fmt.Errorf("invalid value for --enable_safety_checker: '%s'. Must be true or false", flagValue)
				}
				i++ // Consume only the flag=value arg
			} else if i+1 < len(args) && (strings.ToLower(args[i+1]) == "true" || strings.ToLower(args[i+1]) == "false") {
				val, _ = strconv.ParseBool(args[i+1])
				i += 2 // Consume flag and value
			} else {
				val = true // Assume --flag means true
				i++        // Consume only the flag
			}
			parsedReq.EnableSafetyChecker = &val
		case "--safety_tolerance":
			if i+1 < len(args) {
				parsedReq.SafetyTolerance = args[i+1]
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --safety_tolerance argument")
			}
		case "--output_format":
			if i+1 < len(args) {
				parsedReq.OutputFormat = strings.ToLower(args[i+1])
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --output_format argument")
			}
		case "--negative_prompt", "--negative-prompt":
			if i+1 < len(args) {
				parsedReq.NegativePrompt = args[i+1] // Keep original case for prompt
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --negative_prompt argument")
			}
		case "--guidance_scale", "--guidance-scale":
			if i+1 < len(args) {
				val, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil {
					return "", nil, fmt.Errorf("invalid value for --guidance_scale: %s", args[i+1])
				}
				parsedReq.GuidanceScale = &val
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --guidance_scale argument")
			}
		case "--aspect_ratio", "--aspect-ratio":
			if i+1 < len(args) {
				parsedReq.AspectRatio = args[i+1]
				i += 2
			} else {
				return "", nil, fmt.Errorf("missing value for --aspect_ratio argument")
			}
		case "--raw":
			var val bool
			var err error
			if flagValue != "" { // Handle --flag=value
				val, err = strconv.ParseBool(flagValue)
				if err != nil {
					return "", nil, fmt.Errorf("invalid value for --raw: '%s'. Must be true or false", flagValue)
				}
				i++
			} else if i+1 < len(args) && (strings.ToLower(args[i+1]) == "true" || strings.ToLower(args[i+1]) == "false") {
				val, _ = strconv.ParseBool(args[i+1])
				i += 2
			} else {
				val = true // Assume --raw means true
				i++
			}
			parsedReq.Raw = &val
		default:
			// Assume it's part of the prompt
			promptParts = append(promptParts, args[i]) // Use original arg with case preserved
			i++
		}
	}

	prompt := strings.Join(promptParts, " ")
	if prompt == "" {
		return "", nil, fmt.Errorf("please provide a prompt text")
	}

	return prompt, parsedReq, nil
}
