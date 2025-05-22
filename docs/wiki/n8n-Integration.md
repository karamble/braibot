# n8n Integration Guide

This guide explains how to integrate n8n with BraiBot for advanced AI workflows using the `!ai` command.

## Table of Contents

1. [Overview](#overview)
2. [Setup Options](#setup-options)
3. [Webhook Configuration](#webhook-configuration)
4. [Importing Workflows](#importing-workflows)
5. [Error Handling](#error-handling)
6. [Security](#security)
7. [Monitoring](#monitoring)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The `!ai` command in BraiBot uses n8n webhooks to create flexible and powerful AI processing pipelines. This integration allows you to:

- Process messages with custom AI models
- Chain multiple AI operations
- Integrate with external services
- Handle complex workflows

## Setup Options

### 1. n8n.io Cloud Service

Quick setup using the managed cloud service:

1. Register at [n8n.io](https://n8n.io)
2. Create a new workflow
3. Add a webhook trigger
4. Configure authentication
5. Get your webhook URL

### 2. Self-hosted n8n (Recommended)

For privacy and control, use the [local-ai-packaged](https://github.com/coleam00/local-ai-packaged) solution. Please follow the detailed setup guide in the repository's [README.md](https://github.com/coleam00/local-ai-packaged/blob/main/README.md) for installation and configuration instructions. The guide includes:

- System requirements
- Installation steps
- Configuration options
- GPU support setup
- Troubleshooting steps
- Security considerations

This package includes a comprehensive suite of AI and automation tools:

#### Core Components
- **n8n**: Workflow automation platform with 400+ integrations
- **Flowise**: Visual AI workflow builder and deployment tool
- **Open WebUI**: User interface for managing AI models and workflows

#### AI and Language Models
- **Ollama**: Local LLM support with various models
- **Qdrant**: Vector database for AI embeddings and semantic search
- **LangChain**: Framework for building AI applications

#### Data Storage and Management
- **PostgreSQL**: Primary database for data storage
- **Supabase**: Open source Firebase alternative with:
  - Database management
  - Authentication
  - Real-time subscriptions
  - Storage solutions

#### Search and Web Tools
- **SearXNG**: Privacy-focused meta search engine
- **Caddy**: Modern web server with automatic HTTPS for production deployment

#### Additional Features
- **Local File System Access**: Shared folder for file operations
- **GPU Support**: Optional GPU acceleration for AI models
- **Webhook Management**: Built-in webhook handling
- **API Integrations**: Ready-to-use connections to various services

#### System Requirements
- **CPU Profile**: Basic setup for standard processing
- **GPU Profiles**: 
  - NVIDIA GPU support
  - AMD GPU support
- **Memory**: Minimum 8GB RAM recommended
- **Storage**: 20GB+ free space recommended

#### Security Features
- **Authentication**: Built-in user management
- **HTTPS**: Automatic SSL/TLS with Caddy
- **Credential Management**: Secure storage for API keys
- **Access Control**: Role-based permissions

#### Development Tools
- **API Documentation**: Built-in API reference
- **Debug Tools**: Workflow testing and debugging
- **Monitoring**: Performance and usage tracking
- **Logging**: Comprehensive system logs

## Importing Workflows to n8n

BraiBot comes with a collection of pre-built n8n workflows in the `n8n/workflows` directory. These workflows provide various AI-powered capabilities. By configuring and activating the `Decred Assistant BRaiBot` workflow the webhook is enabled that the bot communicates with. You have to configure the authentication credentials for the webhook. It uses generic->header auth->X-BRAIBOT-API-KEY->yourAPIkey

Make sure to configure an API key in the n8n webhook and use that in combination with the production-url in your braibot.conf

### Available Workflows

1. **Decred Assistant BRaiBot**
   - Main workflow for the webhook the BRaiBot communicates with
   - AI Agent with various tools available and easily extensible
   - Handles price, network, and a decred knowledge base vector database
   - Size: 55KB

2. **Decred Knowledge Database**
   - AI RAG workflow with access to a Supabase vector database

3. **CoinMarketCap Integration**
   - Crypto price and market data, using generic->header auth->X-CMC_PRO_API_KEY->yourAPIkey
   - DEX scanning capabilities
   - AI-powered market analysis
   - Sizes: 12KB - 31KB

3. **Specialized Agents**
   - Technical Analysis Agent, using chart-img.com api for chart snapshots
   - Content Creator Agent, using web search and LLM model for summarizing
   - Gmail Email Agent (disconnected by default, only for private use), uses oauth2 auth from google console
   - Google Calendar Agent (disconnected by default, only for private use), uses oauth2 auth from google console
   - Nextcloud Agent (disconnected by default, only for private use), api key from your nextcloud instance, where Web DAV URL is https://yourNextcloud/remote.php/webdav/
   - SearXNG Agent gives the AI Agent web search capabilities, when using the local ai package the API URL is listening on http://searxng:8080

### Importing Workflows

1. **Access n8n Interface**
   - Open your n8n instance
   - Navigate to Workflows
   - Click "Create Workflow"
   - Copy-Paste the workflow json content into the dashboard, hit save
   - Repeat steps for all workflows json files you want to make available to the AI Agent

### Configuring Credentials

Each workflow requires specific credentials to function. You can click on the nodes to add required credentials and store it in n8n's private credential store.

## Security

### Authentication

1. **Webhook Security**
   - Use header-based authentication

2. **Suggested Access Control**
   - Restrict webhook access per ip
   - Monitor usage patterns
   - Log security events

## Troubleshooting

### Common Issues

1. **Webhook Not Receiving**
   - Check webhook URL - `Production URL` is active when workflow is activated, test URL is active and runs only once when hitting `Test workflow` for detailed flow debugging
   - Verify authentication
   - Make sure Decred Assistant BraiBot workflow is activated and running

2. **Processing Errors**
   - Run the webhook in `Test workflow` mode. Will execute once and give ability to follow flow state and input and outputs for each node. Use the `Test URL` in your braibot.conf to hit the test-webhook endpoint.

3. **CURL example BRaiBot sends to n8n webhook**
```
curl -X POST \
  'https://your-webhook-url' \
  -H 'Content-Type: application/json' \
  -H 'X-BRAIBOT-API-KEY: your-api-key' \
  -d '{
    "message": "!ai Hello, how are you?",
    "user": "username"
  }'
  ```
  What the webhook response looks like
  ```
[
  {
    "query": "!ai Hello, how are you?",
    "session_id": "username",
    "bruser": "username"
  },
  {
    "output": "Hello, i am good.",
    "intermediateSteps": [
      {
        "action": {
          "tool": "example tool",
          "toolInput": {
            "input": "example input created by LLM"
          },
          "toolCallId": "call_YV...B0",
          "log": "Invoking \"Example_Tool\" with {\"input\":\"example input\"}\n",
          "messageLog": [
            { ....detailed AI Agent workflow processing and tool flow information here
            }
          ]
        },
        "observation": "more thinking process and tool responses here"
      }
    ]
  }
]
```