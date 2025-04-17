package utils

import (
	"fmt"
	"net/url"
)

// FormatEmbeddedImageMessage formats a message with an embedded image
func FormatEmbeddedImageMessage(altText string, contentType string, base64Data string) string {
	return fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--",
		url.QueryEscape(altText),
		contentType,
		base64Data)
}

// FormatEmbeddedAudioMessage formats a message with embedded audio
func FormatEmbeddedAudioMessage(modelName string, base64Data string) string {
	return fmt.Sprintf("--embed[alt=%s speech,type=audio/ogg,data=%s]--",
		modelName,
		base64Data)
}

// FormatModelHelpMessage formats a help message for a model
func FormatModelHelpMessage(modelName string, description string, priceUSD float64, priceDCR float64, balanceDCR float64, helpDoc string) string {
	return fmt.Sprintf("Model: %s\nDescription: %s\nPrice: $%.2f USD (%.8f DCR)\n\nYour Balance: %.8f DCR\n\n%s",
		modelName,
		description,
		priceUSD,
		priceDCR,
		balanceDCR,
		helpDoc)
}

// FormatModelListMessage formats a message listing available models
func FormatModelListMessage(commandName string, models map[string]interface{}) string {
	modelList := fmt.Sprintf("Available models for %s:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n", commandName)
	for name, model := range models {
		if m, ok := model.(interface{ GetDescription() string }); ok {
			if p, ok := model.(interface{ GetPriceUSD() float64 }); ok {
				modelList += fmt.Sprintf("| %s | %s | $%.2f |\n", name, m.GetDescription(), p.GetPriceUSD())
			}
		}
	}
	return modelList
}
