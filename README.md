# Chat with Goku

A fun desktop application that lets you chat with Goku from Dragon Ball! This application uses the Fyne GUI toolkit and Ollama's local AI capabilities to create an interactive chat experience with Goku's personality.

## Features

- 🎮 Interactive chat interface with Goku's personality
- 💬 Real-time responses using Ollama's AI
- 🎨 Dark theme UI with responsive design
- ⌨️ Keyboard shortcuts for easy messaging
- 🔄 Loading indicators for better user experience

## Prerequisites

- Go 1.16 or higher
- Ollama installed and running locally
- Llama 3.2 model installed in Ollama

## Installation

1. Make sure you have Ollama installed and running:
   ```bash
   # Install Ollama (if not already installed)
   # Start Ollama service
   ollama serve
   ```

2. Install the Llama 3.2 model in Ollama:
   ```bash
   ollama pull llama3.2
   ```

3. Clone this repository:
   ```bash
   git clone https://github.com/Kiryu-Kazuma23/chat-with-goku.git
   cd chat-with-goku
   ```

4. Install dependencies:
   ```bash
   go mod tidy
   ```

## Building and Running

### Windows
Run the build script:
```powershell
.\build.ps1
```

### Manual Build
```bash
go build -o chat-with-goku
```

## Usage

1. Start the application
2. Type your message in the input field
3. Press Enter or click the send button to chat with Goku
4. Enjoy the conversation with Goku's cheerful personality!

## Project Structure

- `main.go` - Main application code
- `build.ps1` - Windows build script
- `go.mod` - Go module dependencies
- `go.sum` - Go module checksums

## Contributing

Feel free to submit issues and enhancement requests!

## License

This project is open source and available under the [MIT License](LICENSE).

## Author

Kiryu-Kazuma23 