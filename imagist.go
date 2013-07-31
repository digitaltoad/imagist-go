package main

import (
  "net/http"
  "github.com/gorilla/mux"
  "fmt"
  "log"
  "image"
  "image/draw"
  "image/color"
  "image/png"
  "image/jpeg"
  // "code.google.com/p/freetype-go/freetype"
  "strconv"
  "os"
)

func PlaceholderHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  format := vars["f"]

  w.Header().Add("Content-Type", fmt.Sprintf("image/%s", format))

  height, err := strconv.Atoi(vars["h"])
  if err != nil{
    height = 100
  }

  width, err := strconv.Atoi(vars["w"])
  if err != nil {
    width = 100
  }

  m := image.NewRGBA(image.Rect(0, 0, height, width))
  gray := color.RGBA{230, 230, 230, 255}
  draw.Draw(m, m.Bounds(), &image.Uniform{gray}, image.ZP, draw.Src)

  switch format {
    case "png":
      png.Encode(w, m)
    case "jpg":
      jpeg.Encode(w, m, nil)
  }
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/{h:[0-9]+}x{w:[0-9]+}.{f:(png|jpg)}", PlaceholderHandler).
    Methods("GET")

  http.Handle("/", r)

  port := os.Getenv("PORT")

  if port == "" {
    port = "9999"
  }

  fmt.Println("Listening on PORT "+port+"...")
  err := http.ListenAndServe(":"+port, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
