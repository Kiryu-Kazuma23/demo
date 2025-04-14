package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system"`
	Template string `json:"template"`
}

// OllamaResponse represents the response structure from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

const gokuPersonality = `You are Goku from Dragon Ball. Respond in Goku's cheerful, friendly, and battle-loving personality.
Key traits to incorporate:
- Use phrases like "Hey there!", "Awesome!", "That sounds like fun!"
- Show enthusiasm for challenges and training
- Be friendly and optimistic
- Occasionally mention food, especially when you're hungry
- Use Goku's casual speech pattern
- Reference martial arts and training
- Show interest in becoming stronger
- Mention your friends like Krillin, Gohan, or Vegeta when relevant
- Use phrases like "Wow!", "Alright!", and "Let's do this!"
- Express excitement about fighting strong opponents
- Be pure-hearted and always willing to help others
Start your first message with: "Hey there! I'm Goku! *puts hand behind head and grins* What's up?"
`

// getAIResponse sends a request to Ollama and returns the response
func getAIResponse(userMessage string) (string, error) {
	url := "http://localhost:11434/api/generate"

	// Prepare the request
	reqBody := OllamaRequest{
		Model:  "tinyllama",
		Prompt: userMessage,
		System: gokuPersonality,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Send the request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response stream
	decoder := json.NewDecoder(resp.Body)
	var fullResponse strings.Builder

	for {
		var response OllamaResponse
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error decoding response: %v", err)
		}

		fullResponse.WriteString(response.Response)

		if response.Done {
			break
		}
	}

	return fullResponse.String(), nil
}

func main() {
	// Create a new application
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	// Create a new window
	window := myApp.NewWindow("Chat with Goku")
	window.SetIcon(theme.MailComposeIcon())

	// Create a chat history area with custom style
	chatHistory := widget.NewTextGrid()
	chatHistory.SetText("Goku: Hey there! I'm Goku! *puts hand behind head and grins* What's up?\n\n")
	chatHistory.ShowLineNumbers = false

	// Create a scroll container for chat history with proper scaling
	historyBox := container.NewVBox(chatHistory)
	scrollContainer := container.NewScroll(historyBox)

	// Wrap scroll container in a max size container to ensure it expands properly
	chatArea := container.NewMax(
		canvas.NewRectangle(theme.BackgroundColor()),
		container.NewPadded(scrollContainer),
	)

	// Create input field with proper styling and handling
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Talk to Goku...")
	input.Wrapping = fyne.TextWrapWord
	input.MultiLine = false

	// Create loading indicator
	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide()

	// Create send button with better styling
	sendButton := widget.NewButtonWithIcon("", theme.MailSendIcon(), nil)
	sendButton.Importance = widget.HighImportance

	// Set up the send action
	sendAction := func() {
		message := input.Text
		if strings.TrimSpace(message) != "" {
			// Add user message to chat history
			currentText := chatHistory.Text()
			chatHistory.SetText(currentText + "You: " + message + "\n\n")

			// Clear input and show loading
			input.SetText("")
			loadingIndicator.Show()
			sendButton.Disable()

			// Get AI response in a goroutine
			go func() {
				response, err := getAIResponse(message)
				if err != nil {
					response = "Oops! Looks like I messed up there! *scratches head* " + err.Error()
				}

				// Update UI in the main thread
				window.Canvas().Refresh(chatHistory)
				chatHistory.SetText(chatHistory.Text() + "Goku: " + response + "\n\n")
				loadingIndicator.Hide()
				sendButton.Enable()
				scrollContainer.ScrollToBottom()
			}()
		}
	}

	// Set the send button action
	sendButton.OnTapped = sendAction

	// Create a better styled input area with loading indicator
	inputArea := container.NewMax(
		canvas.NewRectangle(theme.BackgroundColor()),
		container.NewPadded(
			container.NewBorder(
				loadingIndicator,
				nil, nil,
				container.NewPadded(sendButton),
				input,
			),
		),
	)
	inputArea.Resize(fyne.NewSize(0, 60))

	// Create the main layout with proper spacing and organization
	mainContent := container.NewBorder(
		nil,
		inputArea,
		nil, nil,
		chatArea,
	)

	// Set content and handle window sizing
	window.SetContent(mainContent)

	// Set minimum size constraints
	window.Resize(fyne.NewSize(600, 400))
	window.SetFixedSize(false)

	// Make window responsive to screen size
	screen := window.Canvas().Size()
	idealWidth := screen.Width * 0.7
	idealHeight := screen.Height * 0.8

	if idealWidth < 600 {
		idealWidth = 600
	}
	if idealHeight < 400 {
		idealHeight = 400
	}

	window.Resize(fyne.NewSize(idealWidth, idealHeight))
	window.CenterOnScreen()

	// Add keyboard shortcut for sending messages
	window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyReturn {
			sendAction()
		}
	})

	// Show and run the application
	window.ShowAndRun()
}
