package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	usageTracker "github.com/clashkid155/usage-monitor"
	"github.com/dustin/go-humanize"
	"io"

	"log"
	"net/http"
	"time"
)

var tableBaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
func (m AppModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Table Example")
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Was a key pressed?
	case tea.KeyMsg:
		// Which key was pressed?
		switch msg.String() {
		// Exit if this was the key pressed
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AppModel) View() string {
	return tableBaseStyle.Render(m.table.View()) + "\n"
}

func main() {
	var initModel AppModel
	initModel.currentNetwork = initCurrentNetwork()
	tColumns := []table.Column{
		{Title: "WiFi name (SSID)", Width: 20},
		{Title: "Download", Width: 10},
		{Title: "Upload", Width: 10},
		{Title: "Total Usage", Width: 15},
	}
	var tRows []table.Row

	for _, usage := range initModel.currentNetwork.usages {
		tRows = append(tRows, table.Row{usage.SSID, humanize.Bytes(usage.Download), humanize.Bytes(usage.Upload), humanize.Bytes(usage.TotalUsage)})
	}
	t := table.New(table.WithColumns(tColumns), table.WithRows(tRows), table.WithFocused(true))

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	initModel.table = t
	if _, err := tea.NewProgram(initModel).Run(); err != nil {
		log.Fatalln("Error running program:", err)
	}
}

/*
	func dateFormatter(date string) string  {
		parseTime, err :=time.Parse("02-01-2006", date)
		if err != nil {
			log.Fatal(err)
		}
		return humanize.Time(parseTime)
	}
*/
func dateFormatter(date string) string {
	parseTime, err := time.Parse("02-01-2006", date)
	if err != nil {
		log.Fatal(err)
	}
	return parseTime.Format("02-01-2006")
}
func initCurrentNetwork() *currentNetwork {
	network := new(currentNetwork)
	// usagesByDate := getUsageByDate(time.Now().Format("02-01-2006"))
	usagesByDate := getUsageByDate("05-02-2025")
	if usagesByDate == nil || usagesByDate.Error != "" {
		return network
	}
	network.usages = usagesByDate.Data
	for _, usage := range usagesByDate.Data {
		network.totalUpload += usage.Upload
		network.totalDownload += usage.Download
	}
	network.totalUsage = network.totalUpload + network.totalDownload
	return network
}

func getUsageByDate(date string) *JsonResponse {
	res, err := http.Get(fmt.Sprintf("http://localhost:9083/getUsageByDate?year=%s", date))
	if err != nil {
		log.Println(err, "\nServer not reachable")
		return nil
	}
	defer res.Body.Close()

	var jsonResponse *JsonResponse
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		log.Println(err)
		return nil
	}

	return jsonResponse
}

type JsonResponse struct {
	Message string               `json:"message"`
	Data    []usageTracker.Usage `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}
