# User Guide

Welcome to the BraiBot User Guide! This guide will help you get started with BraiBot and make the most of its features.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Basic Commands](#basic-commands)
4. [Advanced AI Commands](#advanced-ai-commands)
5. [Billing](#billing)
6. [Troubleshooting](#troubleshooting)

## Prerequisites

Before you can use BraiBot, you'll need:

* A Bison Relay account and a running Bison Relay Instance
* Funded Bison Relay LN wallet with open inbound and outbound channel capacity
* Basic understanding of Bison Relay commands

## Installation

Follow these steps to install and configure BraiBot:

1. Set up Bison Relay
   * Install Bison Relay Client
   * Run Bison Relay Client and create an account
   * Fund your account with Decred or use a paid invite key
   * Stop the Client
   * Enable RPC interface in `~/.brclient/brclient.conf`:
     ```ini
     [clientrpc]
     # Enable the JSON-RPC clientrpc protocol
     jsonrpclisten = 127.0.0.1:7676
     
     # TLS certificate paths
     rpccertpath = ~/.brclient/rpc.cert
     rpckeypath = ~/.brclient/rpc.key
     rpcclientcapath = ~/.brclient/rpc-ca.cert
     
     # Generate client certificates automatically
     rpcissueclientcert = true
     ```
   * Restart Bison Relay Client after configuration changes

2. Configure your Lightning Network wallet
   * Ensure your wallet is funded
   * Verify channel capacity for inbound and outbound payments

3. Install BraiBot
   * Clone the repository
   * Build the bot
   * Run the bot for the first time

4. First Launch Configuration
   * The bot automatically detects and configures Bison Relay RPC settings from `~/.brclient/brclient.conf`
   * Follow the first-launch wizard to:
     * Configure your fal.ai API key
     * Enable or disable billing system
     * Configure !ai command capabilities:
       * Enable/disable n8n webhook integration
       * Set up n8n webhook URL (if enabled)
   * The configuration will be saved for future launches

[View n8n integration guide →](n8n-Integration)

## Basic Commands

BraiBot offers several basic commands for AI generation with fal.ai:

* `!text2imgage` - Generate images from text
* `!text2video` - Create videos from text
* `!text2speech` - Convert text to speech
* `!imgage2imge` - Transform images
* `!imgage2video` - Create videos from images

## Advanced AI Commands

For advanced AI Agents with n8n integration, BraiBot offers:

* `!ai` command with n8n webhook integration
* Custom AI workflows
* Run in cloud or as local ai solution

[Explore the n8n Integration Guide →](n8n-integration)

## Billing

BraiBot uses a local balance system with Lightning Network integration for payments. Each user's balance is stored in a local SQLite database and can be managed through various commands. Users can fund their balance trough Bison Relay's tipping feature.

### Balance Management

* **Storage**: Balances are stored in a local SQLite database (`data/balances.db`)
* **Currency**: Balances are stored in DCR atoms (1 DCR = 1e11 atoms)
* **Exchange Rate**: DCR/USD rates are fetched from CoinGecko every 5 minutes
* **Automatic Deduction**: Costs are automatically deducted when commands are executed
* **Balance Check**: System verifies sufficient balance before executing paid commands

### Cost Structure

Each command has an associated cost in USD, which is automatically converted to DCR using current exchange rates. Execution cost can be configured for each model.


### Balance Commands

* `!balance` - Check your current balance

### Free Mode

When billing is disabled (via configuration), all commands are free to use. You'll see a "Happy Days! All commands are free to use" message in the help command.

### Troubleshooting

If you encounter balance-related issues:

* Check your balance with `!balance`
* Ensure you have sufficient funds
* Verify the command cost
* Check if you're in free mode

For more help, see the [Troubleshooting](#troubleshooting) section.

## Troubleshooting

Common issues and solutions:

* Connection problems
* Payment issues
* Generation errors
* Bison Relay specific issues