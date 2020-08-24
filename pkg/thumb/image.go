package thumb

import (
	"errors"
	"fmt"
	"github.com/HFO4/cloudreve/pkg/util"
	im "image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// Thumb 缩略图
type Thumb struct {
	src im.Image
	ext string
}

// NewThumbFromFile 从文件数据获取新的Thumb对象，
// 尝试通过文件名name解码图像
func NewThumbFromFile(file io.Reader, name string) (*Thumb, error) {
	ext := strings.ToLower(filepath.Ext(name))
	// 无扩展名时
	if len(ext) == 0 {
		return nil, errors.New("未知的图像类型")
	}

	var err error
	var img im.Image
	switch ext[1:] {
	case "jpg":
		img, err = jpeg.Decode(file)
	case "jpeg":
		img, err = jpeg.Decode(file)
	case "gif":
		img, err = gif.Decode(file)
	case "png":
		img, err = png.Decode(file)
	default:
		return nil, errors.New("未知的图像类型")
	}
	if err != nil {
		return nil, err
	}

	return &Thumb{
		src: img,
		ext: ext,
	}, nil
}

// GetThumb 生成给定最大尺寸的缩略图
func (image *Thumb) GetThumb(width, height uint) {
	w, h := image.GetSize()
	if w > h {
		ratio := float64(height) / float64(h)
		newWidth := uint(float64(w) * ratio)
		image.src = resize.Resize(newWidth, height, image.src, resize.Lanczos3)
		rectX0 := newWidth/2 - width/2
		rectX1 := newWidth/2 + width/2
		image.src = image.src.(interface{ SubImage(r im.Rectangle) im.Image }).SubImage(im.Rect(int(rectX0), 0, int(rectX1), int(height)))
	} else {
		ratio := float64(width) / float64(w)
		newHeight := uint(float64(h) * ratio)
		image.src = resize.Resize(width, newHeight, image.src, resize.Lanczos3)
		image.src = image.src.(interface{ SubImage(r im.Rectangle) im.Image }).SubImage(im.Rect(0, 0, int(width), int(height)))
	}
	//image.src = resize.Thumbnail(width, height, image.src, resize.Lanczos3)
}

// GetSize 获取图像尺寸
func (image *Thumb) GetSize() (int, int) {
	b := image.src.Bounds()
	return b.Max.X, b.Max.Y
}

// Save 保存图像到给定路径
func (image *Thumb) Save(path string) (err error) {
	out, err := util.CreatNestedFile(path)

	if err != nil {
		return err
	}
	defer out.Close()

	err = png.Encode(out, image.src)
	return err

}

// CreateAvatar 创建头像
func (image *Thumb) CreateAvatar(uid uint, savePath string, s int, m int, l int) error {
	// 生成头像缩略图
	src := image.src
	for k, size := range []int{s, m, l} {
		image.src = resize.Resize(uint(size), uint(size), src, resize.Lanczos3)
		err := image.Save(filepath.Join(savePath, fmt.Sprintf("avatar_%d_%d.png", uid, k)))
		if err != nil {
			return err
		}
	}

	return nil

}
