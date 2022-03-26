package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	chrome   *Chrome
	localDir string
)

func init() {
	width := CONFIG.GetInt32("", "width")
	height := CONFIG.GetInt32("", "height")
	chrome = NewChrome().SetHeight(int(height)).SetWith(int(width))

	localDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
}

func getCode(char string) int64 {

	textQuoted := strconv.QuoteToASCII(char)
	textQuoted = strings.Replace(textQuoted, "\"", "", -1)
	textQuoted = strings.Replace(textQuoted, "\\u", "", -1)
	code, _ := strconv.ParseInt(textQuoted, 16, 32)

	return code
}

func genPictures(char string) {
	code := getCode(char)
	in, err := os.Open(fmt.Sprintf("%s%v.svg", CONFIG.GetString("", "data_dir"), code))
	defer func() {
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}()

	if err != nil {
		panic(err)
	}
	defer in.Close()

	jsonInfo := parseJson(char)

	data, err := ioutil.ReadAll(in)
	xmlData := string(data)

	color := strings.Replace(CONFIG.GetString("", "gif_char_color"), "0x", "#", -1)
	xmlData = strings.Replace(xmlData, "black", color, -1)
	xmlData = strings.Replace(xmlData, "blue", color, -1)

	bgcolor := strings.Replace(CONFIG.GetString("", "gif_char_bg_color"), "0x", "#", -1)
	xmlData = strings.Replace(xmlData, "lightgray", bgcolor, -1)

	compColor := strings.Replace(CONFIG.GetString("", "comp_color"), "0x", "#", -1)

	showGrid := true
	if CONFIG.GetInt32("", "show_grid") == 0 {
		showGrid = false
	}

	out, _ := os.Create("./temp/temp.svg")
	defer out.Close()
	out.WriteString(xmlData)

	reg := regexp.MustCompile(`animation-delay:.*`)
	indices := reg.FindAllStringIndex(xmlData, -1)

	reg2 := regexp.MustCompile(`animation:.*?both`)
	indices2 := reg2.FindAllStringIndex(xmlData, -1)

	regKF := regexp.MustCompile(`@keyframes.*`)
	regStroke := regexp.MustCompile(`stroke:.*`)

	total := len(indices)
	str := strings.Replace(xmlData[indices[total-1][0]:indices[total-1][1]], ";", "", -1)
	str = strings.Trim(strings.Replace(str, "s", "", -1), " ")
	strs := strings.Split(str, ":")
	lastTime, _ := strconv.ParseFloat(strings.Trim(strs[1], " "), 64)

	idx2 := indices2[len(indices2)-1]
	strs2 := strings.Split(strings.Replace(xmlData[idx2[0]:idx2[1]], ";", "", -1), " ")
	duration, _ := strconv.ParseFloat(strings.Replace(strings.Trim(strs2[2], " "), "s", "", -1), 64)

	totalTime := lastTime + duration
	speed := CONFIG.GetFloat64("", "svg_interval")
	count := int64(math.Ceil(totalTime/speed)) + 1

	outDir := CONFIG.GetString("", "out_dir")
	os.MkdirAll(CONFIG.GetString("", "out_dir"), os.ModePerm)

	os.MkdirAll("./temp", os.ModePerm)
	os.MkdirAll(outDir+"/"+char, os.ModePerm)

	for i := 0; int64(i) <= count; i++ {
		tempLines := strings.Split(xmlData, "\n")
		lines := make([]string, 0)

		kfCount := -1
		needRed := false

		for iLine, line := range tempLines {
			if !showGrid && iLine >= 1 && iLine <= 6 {
				continue
			}

			if reg.MatchString(line) {
				str := strings.Replace(line, ";", "", -1)
				strs := strings.Split(str, ":")
				timeStr := strings.Trim(strings.Trim(strs[1], "s"), " ")
				tm, _ := strconv.ParseFloat(timeStr, 64)

				newline := fmt.Sprintf("animation-delay: %fs;", -(float64(i)-1)*speed+tm)

				lines = append(lines, newline)
			} else if regKF.MatchString(line) {
				kfCount++
				needRed = false
				for _, v := range jsonInfo.RadStrokes {
					if kfCount == v {
						needRed = true
						break
					}
				}
				lines = append(lines, line)
			} else if needRed && regStroke.MatchString(line) {
				lines = append(lines, fmt.Sprintf("stroke: %s;", compColor))
			} else {
				lines = append(lines, line)
			}

		}

		svgFilename := localDir + "/temp/temp.svg"
		svgFile, _ := os.Create(svgFilename)
		svgFile.WriteString(strings.Join(lines, "\n"))
		svgFile.Close()

		filepath := fmt.Sprintf("%s/%s%s/%v.png", localDir, outDir, char, i)
		if err := chrome.Screenshoot(svgFilename, filepath, CONFIG.GetString("", "gif_background")); err != nil {
			logrus.Panic(err)
		}
		// exec.Command("open", filepath).Run()
	}

	createGIF(char, int(count), int(CONFIG.GetInt32("", "gif_delay")))
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func genImage(name string) {
	path := fmt.Sprintf("%s/%s%v.svg", localDir, CONFIG.GetString("", "data_dir"), name)
	if ok, _ := PathExists(path); !ok {
		fmt.Printf("%s文件不存在", path)
		return
	}

	outDir := CONFIG.GetString("", "out_dir")
	os.MkdirAll(CONFIG.GetString("", "out_dir"), os.ModePerm)

	filepath := fmt.Sprintf("%s/%s%s.png", localDir, outDir, name)
	if err := chrome.Screenshoot(path, filepath, CONFIG.GetString("", "gif_background")); err != nil {
		logrus.Panic(err)
	}
}

type ImageSet interface {
	Set(x, y int, c color.Color)
}

func createGIF(char string, nframes, delay int) {
	opts := gif.Options{
		NumColors: 256,
		Drawer:    draw.FloydSteinberg,
	}

	anim := gif.GIF{LoopCount: nframes}
	// anim.BackgroundIndex = 0

	outDir := CONFIG.GetString("", "out_dir")
	// bgColor := CONFIG.GetString("", "gif_background")
	// alpah, _ := strconv.ParseInt(bgColor[6:], 16, 8)

	for i := 0; i < nframes; i++ {
		f, err := os.Open(fmt.Sprintf(outDir+char+"/%v.png", i))
		if err != nil {
			fmt.Printf("Could not open file %v.png. Error: %s\n", i, err)
			return
		}
		defer f.Close()
		img, _, _ := image.Decode(f)

		paletted := image.NewPaletted(img.Bounds(), AlphaPalette)
		if opts.Quantizer != nil {
			paletted.Palette = opts.Quantizer.Quantize(make(color.Palette, 0, opts.NumColors), img)
		}
		draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.ZP)

		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, paletted)
	}
	var filename = outDir + char + ".gif"
	file, _ := os.Create(filename)
	defer file.Close()
	gif.EncodeAll(file, &anim)
}
