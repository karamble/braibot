# Braibot

**Braibot** is an AI-powered assistant for the **[Bison Relay](https://bisonrelay.org/)** private messaging platform.

It connects to the **[Fal.ai](https://fal.ai/)** service to let you generate images, videos, and audio directly within your private chats. To pay for the AI services, Braibot uses small, private payments over the **[Decred](https://decred.org/)** Lightning Network.

*(Braibot is built using the helpful [BisonBotKit](https://github.com/vctt94/bisonbotkit) framework).*

## What Can Braibot Do?

*   **AI Image Generation:** Create unique images from text descriptions using various AI models.
*   **AI Image Transformation:** Modify existing images using AI (e.g., apply artistic styles).
*   **AI Video Generation:** Create short video clips from text descriptions or existing images.
*   **AI Text-to-Speech:** Convert your text messages into spoken audio clips using different voices.
*   **Decred Lightning Payments:** Add funds to your bot balance by sending tips via Bison Relay's built-in Decred Lightning Network feature. The bot automatically uses your balance to pay for AI tasks.
*   **Easy Model Selection:** List available AI models for different tasks and choose the one you prefer.
*   **Balance Checking:** Check your current DCR balance with the bot at any time.
*   **Simple Commands:** Interact with the bot using straightforward commands in your private chat.
*   **Helpful Guidance:** Get general help or specific details about commands and AI models.

## What You Need to Use Braibot

1.  **Bison Relay:** You need to have Bison Relay installed and an active account.
2.  **Fal.ai Account & API Key:** Sign up at [Fal.ai](https://fal.ai/), get an API key, and add some credits to pay for the AI generation.
3.  **Decred:** You'll need some Decred (DCR) if you want users (or yourself) to be able to add funds to the bot by sending tips over the Lightning Network.

## Setting Up Braibot

*(These steps are for the person running the bot server).*

1.  **Get the Code:** Download or clone the Braibot code from its repository.
    ```bash
    git clone https://github.com/karamble/braibot.git
    cd braibot
    ```
2.  **Build the Bot:** Compile the bot application.
    ```bash
    go build
    ```
3.  **Configure Bison Relay:** Ensure your Bison Relay client is configured to allow external programs (like Braibot) to connect to it via its RPC interface. This usually involves editing your `brclient.conf` file to enable the `clientrpc` settings (like `jsonrpclisten`, `rpccertpath`, etc.). Refer to Bison Relay documentation for details.
4.  **Configure Braibot:**
    *   The first time you run Braibot, it will try to find your Bison Relay configuration and create its own configuration directory (usually `~/.braibot/`).
    *   It will create a `braibot.conf` file inside that directory.
    *   The bot will likely ask you for your Fal.ai API key during this first run if it's not already in the config file.
    *   You can also manually edit `~/.braibot/braibot.conf` and add your key like this:
        `falapikey=your-fal-ai-api-key`

## Running Braibot

1.  **Start Bison Relay:** Make sure your Bison Relay client is running.
2.  **Start Braibot:** Run the compiled program.
    ```bash
    ./braibot
    ```
    *(You might need to run it from the directory containing the code or provide the full path).*

## Using Braibot (Commands)

Once the bot is running and you've added it as a contact in Bison Relay, send it these commands in a private chat:

*   **`!help`**: Shows the main help message, including your current balance and selected models.
*   **`!help [command]`**: Shows detailed help for a specific command (e.g., `!help text2image`).
*   **`!help [command] [model]`**: Shows details about a specific AI model for a command (e.g., `!help text2image fast-sdxl`).
*   **`!balance`**: Shows your current DCR balance held by the bot. (Add funds by sending tips!).
*   **`!rate`**: Shows the current DCR/USD exchange rate used for pricing AI tasks.
*   **`!listmodels [task]`**: Lists available AI models for a task. Tasks are: `text2image`, `image2image`, `text2speech`, `image2video`, `text2video`.
    *   Example: `!listmodels text2image`
*   **`!setmodel [task] [model_name]`**: Sets the default AI model you want to use for a specific task. Use a model name from `!listmodels`.
    *   Example: `!setmodel text2image fast-sdxl`
*   **`!text2image [your text prompt]`**: Creates an image from your text description using your currently selected text-to-image model.
    *   Example: `!text2image a photo of an astronaut riding a horse on the moon`
*   **`!image2image [image URL] [optional prompt]`**: Transforms the image at the URL using your selected image-to-image model. Some models might use the optional text prompt.
    *   Example: `!image2image https://example.com/photo.jpg turn this into a van gogh painting`
*   **`!image2video [image URL] [optional prompt]`**: Creates a video from the image at the URL using your selected image-to-video model.
    *   Example: `!image2video https://example.com/cat.jpg make the cat slowly blink`
*   **`!text2video [your text prompt]`**: Creates a video from your text description using your selected text-to-video model.
    *   Example: `!text2video cinematic drone shot flying over a futuristic city`
*   **`!text2speech [optional voice ID] [text to speak]`**: Creates an audio clip of the text being spoken. If you don't specify a voice ID, a default voice is used. Check `!help text2speech` for available voice IDs.
    *   Example: `!text2speech Hello from BraiBot!`
    *   Example: `!text2speech Friendly_Person How are you today?`

## Troubleshooting Tips

*   **Bot not responding?** Make sure your Bison Relay client is running and that Braibot is running and connected to it. Check the Braibot logs for connection errors.
*   **Commands failing?**
    *   Check your balance using `!balance`. You might need to send the bot a tip.
    *   Make sure you've entered the command correctly (`!help` is your friend!).
    *   Ensure your Fal.ai account has credits.

## Contributing

(Standard contributing guidelines - Fork, Branch, Commit, Push, Pull Request)

## License

This project uses the ISC License. See the LICENSE file for details.

## Acknowledgments

*   [Bison Relay](https://github.com/companyzero/bisonrelay/)
*   [BisonBotKit](https://github.com/vctt94/bisonbotkit)
*   [Fal.ai](https://fal.ai/)
*   [Decred](https://decred.org/)
