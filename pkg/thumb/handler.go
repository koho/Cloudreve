package thumb

import (
	"bytes"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/h2non/bimg"
	"image"
	"image/jpeg"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

var Handlers = []Handler{NewImageThumb(), NewVideoThumb(), NewPDFThumb()}

func GetLocalSupportedThumbExt() []string {
	ext := make([]string, 0)
	for _, handler := range Handlers {
		ext = append(ext, handler.GetExtension()...)
	}
	return ext
}

type Handler interface {
	GenerateThumb(file io.Reader, name string) (*Thumb, error)
	CanHandle(fileName string) bool
	GetExtension() []string
}

type BaseThumb struct {
	Extension []string
}

func (t *BaseThumb) CanHandle(fileName string) bool {
	ext := getFileExt(fileName)
	return util.ContainsString(t.Extension, ext)
}

func (t *BaseThumb) GetExtension() []string {
	return t.Extension
}

func getFileExt(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
}

type ImageThumb struct {
	*BaseThumb
}

func NewImageThumb() *ImageThumb {
	return &ImageThumb{&BaseThumb{[]string{".jpg", ".jpeg", ".png", ".gif"}}}
}

func (t *ImageThumb) GenerateThumb(file io.Reader, name string) (*Thumb, error) {
	return NewThumbFromFile(file, name)
}

type VideoThumb struct {
	*BaseThumb
}

func NewVideoThumb() *VideoThumb {
	return &VideoThumb{&BaseThumb{[]string{".str", ".aa", ".aac", ".ac3", ".acm", ".adf", ".adp", ".dtk", ".ads", ".ss2", ".adx", ".aea", ".afc", ".aix", ".al", ".ape", ".apl", ".mac", ".aptx", ".aptxhd", ".aqt", ".ast", ".avi", ".avs", ".avr", ".avs", ".avs2", ".bfstm", ".bcstm", ".bit", ".bmv", ".brstm", ".cdg", ".cdxl", ".xl", ".c2", ".302", ".daud", ".str", ".dav", ".dss", ".dts", ".dtshd", ".dv", ".dif", ".cdata", ".eac3", ".paf", ".fap", ".flm", ".flac", ".flv", ".fsb", ".g722", ".722", ".tco", ".rco", ".g723_1", ".g729", ".genh", ".gsm", ".h261", ".h26l", ".h264", ".264", ".avc", ".hevc", ".h265", ".265", ".idf", ".ifv", ".cgi", ".sf", ".ircam", ".ivr", ".kux", ".669", ".amf", ".ams", ".dbm", ".digi", ".dmf", ".dsm", ".dtm", ".far", ".gdm", ".ice", ".imf", ".it", ".j2b", ".m15", ".mdl", ".med", ".mmcmp", ".mms", ".mo3", ".mod", ".mptm", ".mt2", ".mtm", ".nst", ".okt", ".plm", ".ppm", ".psm", ".pt36", ".ptm", ".s3m", ".sfx", ".sfx2", ".st26", ".stk", ".stm", ".stp", ".ult", ".umx", ".wow", ".xm", ".xpk", ".flv", ".lvf", ".m4v", ".mkv", ".mk3d", ".mka", ".mks", ".mjpg", ".mjpeg", ".mpo", ".j2k", ".mlp", ".mov", ".mp4", ".m4a", ".3gp", ".3g2", ".mj2", ".mp2", ".mp3", ".m2a", ".mpa", ".mpc", ".mjpg", ".txt", ".mpl2", ".sub", ".msf", ".mtaf", ".ul", ".musx", ".mvi", ".mxg", ".v", ".nist", ".sph", ".nsp", ".nut", ".ogg", ".oma", ".omg", ".aa3", ".pjs", ".pvf", ".yuv", ".cif", ".qcif", ".rgb", ".rt", ".rsd", ".rsd", ".rso", ".sw", ".sb", ".smi", ".sami", ".sbc", ".msbc", ".sbg", ".scc", ".sdr2", ".sds", ".sdx", ".ser", ".shn", ".vb", ".son", ".sln", ".mjpg", ".stl", ".sub", ".sub", ".sup", ".svag", ".tak", ".thd", ".tta", ".ans", ".art", ".asc", ".diz", ".ice", ".nfo", ".txt", ".vt", ".ty", ".ty+", ".uw", ".ub", ".v210", ".yuv10", ".vag", ".vc1", ".rcv", ".viv", ".idx", ".vpk", ".txt", ".vqf", ".vql", ".vqe", ".vtt", ".wsd", ".xmv", ".xvag", ".yop", ".y4m"}}}
}

func (t *VideoThumb) GenerateThumb(file io.Reader, name string) (*Thumb, error) {
	var err error
	var img image.Image
	cmd := exec.Command("ffmpeg", "-i", name, "-ss", "3", "-vf", "select=gt(scene\\,0.5)", "-vframes", "1", "-vsync", "vfr", "-f", "singlejpeg", "-")
	cmd.Stdin = file
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	img, err = jpeg.Decode(&buffer)
	if err != nil {
		return nil, err
	}
	return &Thumb{
		src: img,
		ext: getFileExt(name),
	}, nil
}

type PDFThumb struct {
	*BaseThumb
}

func NewPDFThumb() *PDFThumb {
	return &PDFThumb{&BaseThumb{[]string{".pdf"}}}
}

func (t *PDFThumb) GenerateThumb(file io.Reader, name string) (*Thumb, error) {
	options := bimg.Options{
		Gravity: bimg.GravityNorth,
		Width:   300,
		Height:  285,
		Crop:    true,
		Quality: 100,
	}
	buffer, err := bimg.Read(name)
	if err != nil {
		return nil, err
	}

	newImage, err := bimg.NewImage(buffer).Convert(bimg.JPEG)
	if err != nil {
		return nil, err
	}
	newImage, err = bimg.NewImage(newImage).Process(options)
	img, err := jpeg.Decode(bytes.NewReader(newImage))
	if err != nil {
		return nil, err
	}
	return &Thumb{
		src: img,
		ext: getFileExt(name),
	}, nil
}

type OfficeThumb struct {
	*BaseThumb
}

func NewOfficeThumb() *OfficeThumb {
	return &OfficeThumb{&BaseThumb{[]string{".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx"}}}
}

func (t *OfficeThumb) GenerateThumb(file io.Reader, name string) (*Thumb, error) {
	return nil, nil
}
