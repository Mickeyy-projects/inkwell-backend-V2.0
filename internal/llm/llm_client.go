package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	ollamaURL string
	client    *http.Client
}

func NewOllamaClient(url string) *OllamaClient {
	return &OllamaClient{
		ollamaURL: url,
		client: &http.Client{
			Timeout: 30 * time.Second, // Set a timeout to avoid hanging requests
		},
	}
}

func (o *OllamaClient) callOllama(prompt string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":  "mistral",
		"prompt": prompt,
	})

	req, err := http.NewRequest("POST", o.ollamaURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Log the full response for debugging
	log.Println("Full LLM response body:", string(bodyBytes))

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if responseText, ok := result["response"].(string); ok {
		return responseText, nil
	}

	return "", errors.New("invalid response from Ollama")
}

func (o *OllamaClient) GenerateQuestions(topic string, limit int) ([]string, error) {
	prompt := fmt.Sprintf("Generate %d multiple-choice questions on %s.", limit, topic)
	response, err := o.callOllama(prompt)
	if err != nil {
		return nil, err
	}
	return parseQuestions(response), nil
}

func (o *OllamaClient) EvaluateAnswer(question, userAnswer, correctAnswer string) (bool, string, error) {
	prompt := fmt.Sprintf("Question: %s\nUser Answer: %s\nCorrect Answer: %s\nIs the answer correct? Explain why.", question, userAnswer, correctAnswer)
	response, err := o.callOllama(prompt)
	if err != nil {
		return false, "", err
	}
	isCorrect := determineCorrectness(response)
	return isCorrect, response, nil
}

func parseQuestions(response string) []string {
	var questions []string
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if line != "" {
			questions = append(questions, line)
		}
	}
	return questions
}

func determineCorrectness(response string) bool {
	return !strings.Contains(strings.ToLower(response), "incorrect")
}

func (o *OllamaClient) CorrectSentence(sentence string) (string, string, error) {
	prompt := "Please correct the following sentence if needed and provide feedback in the format 'Corrected: <corrected sentence> Feedback: <feedback message>': " + sentence
	response, err := o.callOllama(prompt)
	if err != nil {
		return sentence, "Could not generate feedback", err
	}
	log.Println("LLM response:", response)
	// Parse response expecting: "Corrected: <corrected sentence> Feedback: <feedback message>"
	parts := strings.Split(response, "Feedback:")
	var correctedText, feedback string
	if len(parts) >= 2 {
		correctedPart := strings.TrimSpace(parts[0])
		if strings.HasPrefix(correctedPart, "Corrected:") {
			correctedText = strings.TrimSpace(strings.TrimPrefix(correctedPart, "Corrected:"))
		} else {
			correctedText = sentence
		}
		feedback = strings.TrimSpace(parts[1])
	} else {
		correctedText = sentence
		feedback = "No feedback provided"
	}
	return correctedText, feedback, nil
}
