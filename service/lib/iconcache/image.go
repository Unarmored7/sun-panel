package iconcache

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"strings"
)

func prepare(data []byte) ([]byte, string, error) {
	if len(data) >= 6 && bytes.Equal(data[:4], []byte{0, 0, 1, 0}) {
		// ICO frames retain browser support and avoid converting a 16px favicon to a blurry bitmap.
		count := int(data[4]) | int(data[5])<<8
		if count < 1 || count > 100 || len(data) < 6+16*count {
			return nil, "", errors.New("invalid icon")
		}
		for i := 0; i < count; i++ {
			frame := data[6+i*16 : 6+(i+1)*16]
			size := uint64(binary.LittleEndian.Uint32(frame[8:12]))
			offset := uint64(binary.LittleEndian.Uint32(frame[12:16]))
			if size == 0 || offset < uint64(6+16*count) || offset+size > uint64(len(data)) {
				return nil, "", errors.New("invalid ICO frame")
			}
		}
		return data, ".ico", nil
	}
	if bytes.Contains(data[:min(len(data), 1024)], []byte("<svg")) {
		decoder := xml.NewDecoder(bytes.NewReader(data))
		var out bytes.Buffer
		encoder := xml.NewEncoder(&out)
		depth := 0
		root := false
		allowed := map[string]bool{"svg": true, "g": true, "path": true, "circle": true, "rect": true, "ellipse": true, "line": true, "polyline": true, "polygon": true, "defs": true, "linearGradient": true, "radialGradient": true, "stop": true, "use": true, "symbol": true, "clipPath": true, "mask": true, "pattern": true, "title": true, "desc": true, "style": true}
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, "", err
			}
			switch t := token.(type) {
			case xml.StartElement:
				if !allowed[t.Name.Local] || depth >= 64 {
					return nil, "", errors.New("unsupported SVG element")
				}
				if depth == 0 {
					if root || t.Name.Local != "svg" {
						return nil, "", errors.New("invalid SVG root")
					}
					root = true
				}
				for _, a := range t.Attr {
					n := strings.ToLower(a.Name.Local)
					v := strings.ToLower(strings.TrimSpace(a.Value))
					if strings.HasPrefix(n, "on") || (n == "href" && !strings.HasPrefix(v, "#")) || strings.Contains(v, "javascript:") || strings.Contains(v, "@import") || strings.Contains(v, "expression(") {
						return nil, "", errors.New("unsafe SVG attribute")
					}
				}
				attrs := t.Attr[:0]
				for _, a := range t.Attr {
					if a.Name.Local != "xmlns" && a.Name.Space != "xmlns" {
						attrs = append(attrs, a)
					}
				}
				t.Attr = attrs
				token = t
				depth++
			case xml.EndElement:
				depth--
			case xml.Directive, xml.ProcInst:
				continue
			}
			if err := encoder.EncodeToken(token); err != nil {
				return nil, "", err
			}
		}
		if !root || depth != 0 {
			return nil, "", errors.New("invalid SVG")
		}
		encoder.Flush()
		return out.Bytes(), ".svg", nil
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || int64(config.Width)*int64(config.Height) > 16000000 {
		return nil, "", errors.New("unsupported image")
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	width, height := config.Width, config.Height
	if width > 192 || height > 192 {
		scale := 192.0 / float64(max(width, height))
		width = max(1, int(float64(width)*scale))
		height = max(1, int(float64(height)*scale))
	}
	target := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(target, target.Bounds(), source, source.Bounds(), draw.Src, nil)
	var output bytes.Buffer
	if err := png.Encode(&output, target); err != nil {
		return nil, "", err
	}
	return output.Bytes(), ".png", nil
}
