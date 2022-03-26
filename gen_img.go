package main

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	_ "image"
	"io/ioutil"
	"os"
	"path/filepath"

	"os/exec"

	svg "github.com/ajstarks/svgo"
	"github.com/sirupsen/logrus"
)

type CharInfo struct {
	Strokes    []string  `json:"strokes"`
	Medians    [][][]int `json:"medians"`
	RadStrokes []int     `json:"radStrokes"`
}

// {"strokes": ["M 228 673 Q 353 439", "Q 426 512 423 507", "M 633 367"], "medians": [[[512, 332], [522, 342], [584, 355], [642, 347]], [[745, 38], [693, 55], [629, 89]]], "radStrokes": [0, 1]}
func parseJson(char string) (info *CharInfo) {
	jsonDir := CONFIG.GetString("", "json_data_dir")
	in, err := os.Open(jsonDir + char + ".json")
	defer func() {
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}()

	if err != nil {
		panic(err)
	}
	defer in.Close()

	data, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}

	info = new(CharInfo)
	err = json.Unmarshal(data, info)
	if err != nil {
		panic(err)
	}

	return info
}

func parseJsonToSVG(char string, width, height int) (f *os.File, err error) {
	defer func() {
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}()

	info := parseJson(char)

	f, err = os.Create("temp.svg")
	if err != nil {
		panic(err)
	}
	canvas := svg.New(f)

	// canvas.Start(width, height)
	canvas.StartviewUnit(width, height, "", 0, 0, 1200, 1200)
	canvas.ScaleXY(1, -1)

	for _, path := range info.Strokes {
		// canvas.Style("stroke", "#660000")
		// canvas.Style("fill", "none")
		canvas.Path(path, "style=\"stroke:#660000; fill:#ff0000;\"")
	}

	canvas.End()

	return f, err
}

func genSvgImg() {
	chrome := NewChrome().SetHeight(150).SetWith(150)
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	filepath := p + "/out.png"
	if err := chrome.Screenshoot(p+"/阿.svg", filepath, "00000000"); err != nil {
		logrus.Panic(err)
	}

	exec.Command("open", filepath).Run()
}
