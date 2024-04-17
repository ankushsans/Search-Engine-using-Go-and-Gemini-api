package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
	"crypto/tls"
    "sync"
	"strings"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)
func init() {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    if query == "" {
        http.Error(w, "No search query provided", http.StatusBadRequest)
        return
    }

    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
	fmt.Println("go routine called")
        defer wg.Done()
        resp ,err := performSearch(query)
        if err != nil {
            http.Error(w, "Failed to fetch search results", http.StatusInternalServerError)
            return
        }
	   htmlContent := "<h1>Search Results for '" + query + "'</h1><ul>"
        	for _, cand := range resp.Candidates {
            		for _, part := range cand.Content.Parts {
 				switch p := part.(type) {
        			case genai.Text:
					eachlink := strings.Split(string(p),",")
					for i:=0;i<len(eachlink);i++{
					htmlContent += "<li><a href='" + eachlink[i] + "'>" + eachlink[i] + "</a></li>"
				}
        			default:
            			fmt.Println("Unknown part type")

			}  
            }
        }
        htmlContent += "</ul>"

        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, htmlContent)
}()

    wg.Wait()
}

func performSearch(query string) (*genai.GenerateContentResponse, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
	log.Fatal(err)
	}
    	defer client.Close()

  
	model := client.GenerativeModel("gemini-pro")
resp, err := model.GenerateContent(ctx,genai.Text("Give proper and complete URL hyperlinks to all the websites separated by commas, without numbering them ,with the keyword:"),genai.Text(query))

	if err != nil {
	  log.Fatal(err)
	}
    return resp,nil
}
func main() {

	apiKey := "YOUR_API_KEY"
    os.Setenv("API_KEY", apiKey)
    http.Handle("/", http.FileServer(http.Dir(".")))

    http.HandleFunc("/search", searchHandler)

    port := ":8080"
    fmt.Printf("Server listening on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}
