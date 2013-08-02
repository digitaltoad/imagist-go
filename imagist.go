package main

import (
  "flag"
  "net/http"
  "github.com/gorilla/mux"
  "fmt"
  "log"
  "image"
  "image/draw"
  "image/color"
  "image/png"
  "image/jpeg"
  "code.google.com/p/freetype-go/freetype"
  "code.google.com/p/freetype-go/freetype/truetype"
  "io/ioutil"
  "strconv"
  "math"
  "time"
)

var (
  port = flag.String("port", "9999", "web server port")
  fontFile = flag.String("fontFile", "coolvetica.ttf", "filename of the ttf font")
  dpi float64 = 72
)

// HexModel converts any color.Color to an Hex color.
var HexModel = color.ModelFunc(hexModel)

// Hex represents an RGB color in hexadecimal format.
//
// The length must be 3 or 6 characters, preceded or not by a '#'.
type Hex string

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the Hex.
func (c Hex) RGBA() (uint32, uint32, uint32, uint32) {
  r, g, b := HexToRGB(c)
  return uint32(r) * 0x101, uint32(g) * 0x101, uint32(b) * 0x101, 0xffff
}

// hexModel converts a color.Color to Hex.
func hexModel(c color.Color) color.Color {
  if _, ok := c.(Hex); ok {
    return c
  }
  r, g, b, _ := c.RGBA()
  return RGBToHex(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// RGBToHex converts an RGB triple to an Hex string.
func RGBToHex(r, g, b uint8) Hex {
  return Hex(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// HexToRGB converts an Hex string to a RGB triple.
func HexToRGB(h Hex) (uint8, uint8, uint8) {
  if len(h) > 0 && h[0] == '#' {
    h = h[1:]
  }
  if len(h) == 3 {
    h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
  }
  if len(h) == 6 {
    if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
      return uint8(rgb >> 16), uint8((rgb >> 8) & 0xFF), uint8(rgb & 0xFF)
    }
  }
  return 0, 0, 0
}

type Placeholder struct {
  Height int
  Width int
  Image draw.Image
}

func (p *Placeholder) GetFont() (*truetype.Font, error) {
  fontBytes, err := ioutil.ReadFile(*fontFile)
  if err != nil {
    return nil, err
  }

  font, err := freetype.ParseFont(fontBytes)
  if err != nil {
    return nil, err
  }

  return font, nil
}

func (p *Placeholder) GenerateImage() error {
  m := image.NewRGBA(image.Rect(0, 0, p.Height, p.Width))

  hex := Hex("#EEE")
  r, g, b, a := hex.RGBA()
  bg := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
  draw.Draw(m, m.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

  font, err := p.GetFont()
  if err != nil {
    return err
  }

  hex = Hex("#888")
  r, g, b, a = hex.RGBA()
  fg := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

  c := freetype.NewContext()
  c.SetDPI(dpi)
  c.SetFont(font)
  c.SetFontSize(p.GetFontSize())
  c.SetClip(m.Bounds())
  c.SetDst(m)
  c.SetSrc(&image.Uniform{fg})

  text := string(strconv.Itoa(p.Height) + "x" + strconv.Itoa(p.Width))

  pt := freetype.Pt(10, int(p.GetFontSize()))
  _, err = c.DrawString(text, pt)
  if err != nil {
    return err
  }

  p.Image = m

  return nil
}

func (p *Placeholder) GetFontSize() float64 {
  fontPercent := 0.10
  size := 12.0

  if p.Height > p.Width {
    size = float64(p.Width) * fontPercent
  } else {
    size = float64(p.Height) * fontPercent
  }

  return math.Ceil(size)
}

func PlaceholderHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  format := vars["f"]

  height, err := strconv.Atoi(vars["h"])
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  width, err := strconv.Atoi(vars["w"])
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  if height > 2000 || width > 2000 {
    http.Error(w, "Image too big!", http.StatusBadRequest)
    return
  }

  m := Placeholder{Height: height, Width: width}
  err = m.GenerateImage()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Add("Content-Type", fmt.Sprintf("image/%s", format))
  w.Header().Add("Cache-Control", "public, max-age=604800")
  w.Header().Add("Expires", time.Now().AddDate(0, 0, 1).UTC().Format(http.TimeFormat))

  switch format {
    case "png":
      png.Encode(w, m.Image)
    case "jpg":
      jpeg.Encode(w, m.Image, nil)
  }
}

func main() {
  flag.Parse()
  r := mux.NewRouter()
  r.HandleFunc("/{h:[0-9]+}x{w:[0-9]+}.{f:(png|jpg)}", PlaceholderHandler).
    Methods("GET")

  r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

  http.Handle("/", r)

  fmt.Println("Listening on PORT " + *port + "...")

  err := http.ListenAndServe(":" + *port, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
