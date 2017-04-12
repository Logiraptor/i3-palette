package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Info struct {
		Colors [5]string `json:"colors"`
	} `json:"info"`
}

type Theme struct {
	Background string
	StatusLine string
	Separator  string
	Focused    StateTheme
	Active     StateTheme
	Inactive   StateTheme
	Urgent     StateTheme
}

type StateTheme struct {
	Border, Background, Font string
}

func main() {
	imgData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := http.PostForm("http://pictaculous.com/api/1.0/", url.Values{
		"image": {string(imgData)},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var result Response
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println(result.Info.Colors)
	theme := Theme{
		Active: StateTheme{
			Background: result.Info.Colors[0],
			Border:     result.Info.Colors[1],
			Font:       result.Info.Colors[2],
		},
		Focused: StateTheme{
			Background: result.Info.Colors[3],
			Border:     result.Info.Colors[4],
			Font:       result.Info.Colors[0],
		},
		Inactive: StateTheme{
			Background: result.Info.Colors[1],
			Border:     result.Info.Colors[2],
			Font:       result.Info.Colors[3],
		},
		Urgent: StateTheme{
			Background: result.Info.Colors[4],
			Border:     result.Info.Colors[0],
			Font:       result.Info.Colors[1],
		},
		Background: result.Info.Colors[2],
		Separator:  result.Info.Colors[3],
		StatusLine: result.Info.Colors[4],
	}

	template.Must(template.New("root").Parse(configTmpl)).
		Execute(os.Stdout, theme)
}

const configTmpl = `
bar {
    colors {
        # Whole color settings
        background #{{.Background}}
        statusline #{{.StatusLine}}
        separator  #{{.Separator}}

        # Type             border  background font
        focused_workspace  #{{.Focused.Border}} #{{.Focused.Background}} #{{.Focused.Font}}
        active_workspace   #{{.Active.Border}} #{{.Active.Background}} #{{.Active.Font}}
        inactive_workspace #{{.Inactive.Border}} #{{.Inactive.Background}} #{{.Inactive.Font}}
        urgent_workspace   #{{.Urgent.Border}} #{{.Urgent.Background}} #{{.Urgent.Font}}
    }

    status_command i3status
    mode hide
}

# class                 border  backgr. text    indicator child_border
client.focused          #{{.Focused.Border}} #{{.Focused.Background}} #{{.Focused.Font}} #{{.StatusLine}} #{{.Focused.Background}}
client.focused_inactive #{{.Inactive.Border}} #{{.Inactive.Background}} #{{.Inactive.Font}} #{{.StatusLine}} #{{.Inactive.Background}}
client.unfocused        #{{.Inactive.Border}} #{{.Inactive.Background}} #{{.Inactive.Font}} #{{.StatusLine}} #{{.Inactive.Background}}
client.urgent           #{{.Urgent.Border}} #{{.Urgent.Background}} #{{.Urgent.Font}} #{{.StatusLine}} #{{.Urgent.Background}}
client.placeholder      #{{.Urgent.Border}} #{{.Urgent.Background}} #{{.Urgent.Font}} #{{.StatusLine}} #{{.Urgent.Background}}

client.background       #{{.Background}}
`
