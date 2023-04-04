package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
)

var openaiKey string

func main() {
	rootCmd := &cobra.Command{
		Use:   "openai-cli",
		Short: "A command-line tool for OpenAI API",
	}
	dialogueCmd := &cobra.Command{
		Use:   "dialogue",
		Short: "Generate text using OpenAI's GPT-3 language model",
		Run:   dialogue,
	}

	imageRecognitionCmd := &cobra.Command{
		Use:   "image-recognition [filename]",
		Short: "Describe an image using OpenAI's DALL-E image recognition model",
		Args:  cobra.ExactArgs(1),
		Run:   imageRecognition,
	}

	imageGenerationCmd := &cobra.Command{
		Use:   "image-generation [prompt]",
		Short: "Generate an image using OpenAI's DALL-E image generation model",
		Args:  cobra.ExactArgs(1),
		Run:   imageGeneration,
	}

	imageEditingCmd := &cobra.Command{
		Use:   "image-editing [input file] [instructions] [output file]",
		Short: "Edit an image using OpenAI's DALL-E image editing model",
		Args:  cobra.ExactArgs(3),
		Run:   imageEditing,
	}

	audioGenerationCmd := &cobra.Command{
		Use:   "audio-generation [text] [output file]",
		Short: "Generate audio using OpenAI's Jukebox music model",
		Args:  cobra.ExactArgs(2),
		Run:   audioGeneration,
	}

	audioTranscriptionCmd := &cobra.Command{
		Use:   "audio-transcription [filename]",
		Short: "Transcribe speech from an audio file using OpenAI's GPT-3 language model",
		Args:  cobra.ExactArgs(1),
		Run:   audioTranscription,
	}

	codeGenerationCmd := &cobra.Command{
		Use:   "code-generation [prompt]",
		Short: "Generate code using OpenAI's Codex model",
		Args:  cobra.ExactArgs(1),
		Run:   codeGeneration,
	}

	rootCmd.PersistentFlags().StringVar(&openaiKey, "key", "", "OpenAI API key (can also be set using OPENAI_API_KEY environment variable)")

	rootCmd.AddCommand(dialogueCmd)
	rootCmd.AddCommand(imageRecognitionCmd)
	rootCmd.AddCommand(imageGenerationCmd)
	rootCmd.AddCommand(imageEditingCmd)
	rootCmd.AddCommand(audioGenerationCmd)
	rootCmd.AddCommand(audioTranscriptionCmd)
	rootCmd.AddCommand(codeGenerationCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
func getClient() *openai.Client {
	key := openaiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" {
		log.Fatal("OpenAI API key not found. Set it using --key or OPENAI_API_KEY environment variable")
	}
	return openai.NewClient(key)
}
func strToImageBytes(input string) ([]byte, error) {
	// Example JSON input with base64-encoded image data

	// Decode the JSON input into an ImageData object
	type ImageData struct {
		Data string `json:"data"`
	}
	var imageData ImageData
	err := json.Unmarshal([]byte(input), &imageData)
	if err != nil {
		return nil, err
	}

	// Decode the base64-encoded image data into a byte slice
	decoded, err := base64.StdEncoding.DecodeString(imageData.Data)
	if err != nil {
		return nil, err
	}

	// Decode the byte slice into an image object
	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}

	// Convert the image object to a byte slice
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func dialogue(cmd *cobra.Command, args []string) {
	client := getClient()
	prompt := args[0]

	completion, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Prompt:    prompt,
		Model:     "text-davinci-003",
		MaxTokens: 5000,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, choice := range completion.Choices {
		fmt.Println(choice.Text)
	}

}
func imageRecognition(cmd *cobra.Command, args []string) {
	client := getClient()
	filename := args[0]

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	imageBase64 := base64.StdEncoding.EncodeToString(data)

	imageRecogRequest := openai.CompletionRequest{
		Prompt:      fmt.Sprintf("描述这个图片: %s", imageBase64),
		Model:       "image-alpha-001",
		MaxTokens:   50,
		Temperature: 0.5,
		N:           1,
	}

	result, err := client.CreateCompletion(context.Background(), imageRecogRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Choices[0].Text)

}
func imageGeneration(cmd *cobra.Command, args []string) {
	client := getClient()
	prompt := args[0]

	result, err := client.CreateImage(context.Background(), openai.ImageRequest{
		Prompt:         prompt,
		N:              1,
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		User:           "Developer",
	})
	if err != nil {
		log.Fatal("no result, no image data")
	}

	if len(result.Data) == 0 {
		log.Fatal()
	}

	imageBytes, err := strToImageBytes(result.Data[0].B64JSON)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("output.png", imageBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image saved to output.png")

}
func imageEditing(cmd *cobra.Command, args []string) {
	client := getClient()
	inputFile := args[0]
	instructions := args[1]
	outputFile := args[2]

	inputData, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inputData.Close()

	result, err := client.CreateEditImage(context.Background(), openai.ImageEditRequest{
		Image:  inputData,
		Prompt: instructions,
		N:      1,
		Size:   openai.CreateImageSize256x256,
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(result.Data) == 0 {
		log.Fatal("no result, no image edited")
	}

	ret, err := strToImageBytes(result.Data[0].B64JSON)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(outputFile, ret, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Image saved to %s\n", outputFile)

}
func audioGeneration(cmd *cobra.Command, args []string) {
	client := getClient()
	text := args[0]
	outputFile := args[1]

	result, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Prompt:      text,
		Model:       "whisper-3",
		N:           1,
		Temperature: 0.5,
		MaxTokens:   1024,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(outputFile, []byte(result.Choices[0].Text), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Audio saved to %s\n", outputFile)

}
func audioTranscription(cmd *cobra.Command, args []string) {
	client := getClient()
	filename := args[0]

	result, err := client.CreateTranscription(context.Background(), openai.AudioRequest{
		FilePath:    filename,
		Model:       "whisper-3",
		Prompt:      "用简体中文",
		Temperature: 0.5,
		Language:    "zh",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Text)

}
func codeGeneration(cmd *cobra.Command, args []string) {
	client := getClient()
	prompt := args[0]

	result, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   5000,
		Model:       "davinci-codex-002",
		N:           1,
		Temperature: 0.5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Choices[0].Text)

}
