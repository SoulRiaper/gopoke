package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type Stat struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type StatInfo struct {
	Stat     Stat  `json:"stat"`
	BaseStat int32 `json:"base_stat"`
}

type Sprites struct {
	FrontDefault string `json:"front_default"`
	BackDefault  string `json:"back_default"`
}

type Pokemon struct {
	Name     string     `json:"name"`
	BaseExp  int32      `json:"base_experience"`
	Height   int32      `json:"height"`
	Id       int32      `json:"id"`
	Sprites  Sprites    `json:"sprites"`
	StatInfo []StatInfo `json:"stats"`
}

func fetchData(url string, wg *sync.WaitGroup, resultChan chan<- []byte, errorChan chan<- error) {
	defer wg.Done()

	client := http.DefaultClient
	resp, err := client.Get(url)
	if err != nil {
		errorChan <- fmt.Errorf("HTTP request error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorChan <- fmt.Errorf("error reading response body: %v", err)
		return
	}

	resultChan <- body
}

func parseJSON(body []byte) (Pokemon, error) {
	var data Pokemon
	err := json.Unmarshal(body, &data)
	if err != nil {
		return Pokemon{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	return data, nil
}

func downloadSprite(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading sprite: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	spriteData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading sprite data: %v", err)
	}

	return spriteData, nil
}

func saveSprite(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error saving sprite: %v", err)
	}

	return nil
}

func main() {
	url := "https://pokeapi-proxy.freecodecamp.rocks/api/pokemon/1/"

	var wg sync.WaitGroup
	resultChan := make(chan []byte)
	errorChan := make(chan error)

	wg.Add(1)
	go fetchData(url, &wg, resultChan, errorChan)

	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	select {
	case err := <-errorChan:
		fmt.Println("Error:", err)
	case body := <-resultChan:
		pokemon, err := parseJSON(body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Pokemon Name:", pokemon.Name)
		fmt.Println("Pokemon BaseExp:", pokemon.BaseExp)
		fmt.Println("Pokemon Height:", pokemon.Height)
		fmt.Println("Pokemon Id:", pokemon.Id)
		fmt.Println("Pokemon Sprites:", pokemon.Sprites)
		fmt.Println("Pokemon Abilities:", pokemon.StatInfo)

		// Download and save sprites
		if pokemon.Sprites.FrontDefault != "" {
			frontFilename := filepath.Join(".", fmt.Sprintf("%s_front.png", pokemon.Name))
			spriteData, err := downloadSprite(pokemon.Sprites.FrontDefault)
			if err != nil {
				fmt.Println("Error downloading front sprite:", err)
			} else {
				err = saveSprite(spriteData, frontFilename)
				if err != nil {
					fmt.Println("Error saving front sprite:", err)
				} else {
					fmt.Println("Front sprite saved as:", frontFilename)
				}
			}
		}

		if pokemon.Sprites.BackDefault != "" {
			backFilename := filepath.Join(".", fmt.Sprintf("%s_back.png", pokemon.Name))
			spriteData, err := downloadSprite(pokemon.Sprites.BackDefault)
			if err != nil {
				fmt.Println("Error downloading back sprite:", err)
			} else {
				err = saveSprite(spriteData, backFilename)
				if err != nil {
					fmt.Println("Error saving back sprite:", err)
				} else {
					fmt.Println("Back sprite saved as:", backFilename)
				}
			}
		}
	}
}
