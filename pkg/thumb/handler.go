package thumb

import (
	"bytes"
	"github.com/HFO4/cloudreve/pkg/util"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var Handlers = []Handler{NewImageThumb(), NewVideoThumb(), NewPDFThumb(), NewDocThumb()}

func GetLocalSupportedThumbExt() []string {
	ext := make([]string, 0)
	for _, handler := range Handlers {
		ext = append(ext, handler.GetExtension()...)
	}
	return ext
}

type Handler interface {
	GenerateThumb(file io.Reader, name string, url string) (*Thumb, error)
	CanHandle(fileName string) bool
	NeedURL() bool
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

func (t *BaseThumb) NeedURL() bool {
	return false
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

func (t *ImageThumb) GenerateThumb(file io.Reader, name string, url string) (*Thumb, error) {
	return NewThumbFromFile(file, name)
}

type VideoThumb struct {
	*BaseThumb
}

func NewVideoThumb() *VideoThumb {
	return &VideoThumb{&BaseThumb{[]string{".str", ".aa", ".aac", ".ac3", ".acm", ".adf", ".adp", ".dtk", ".ads", ".ss2", ".adx", ".aea", ".afc", ".aix", ".al", ".ape", ".apl", ".mac", ".aptx", ".aptxhd", ".aqt", ".ast", ".avi", ".avs", ".avr", ".avs", ".avs2", ".bfstm", ".bcstm", ".bit", ".bmv", ".brstm", ".cdg", ".cdxl", ".xl", ".c2", ".302", ".daud", ".str", ".dav", ".dss", ".dts", ".dtshd", ".dv", ".dif", ".cdata", ".eac3", ".paf", ".fap", ".flm", ".flac", ".flv", ".fsb", ".g722", ".722", ".tco", ".rco", ".g723_1", ".g729", ".genh", ".gsm", ".h261", ".h26l", ".h264", ".264", ".avc", ".hevc", ".h265", ".265", ".idf", ".ifv", ".cgi", ".sf", ".ircam", ".ivr", ".kux", ".669", ".amf", ".ams", ".dbm", ".digi", ".dmf", ".dsm", ".dtm", ".far", ".gdm", ".ice", ".imf", ".it", ".j2b", ".m15", ".mdl", ".med", ".mmcmp", ".mms", ".mo3", ".mod", ".mptm", ".mt2", ".mtm", ".nst", ".okt", ".plm", ".ppm", ".psm", ".pt36", ".ptm", ".s3m", ".sfx", ".sfx2", ".st26", ".stk", ".stm", ".stp", ".ult", ".umx", ".wow", ".xm", ".xpk", ".flv", ".lvf", ".m4v", ".mkv", ".mk3d", ".mka", ".mks", ".mjpg", ".mjpeg", ".mpo", ".j2k", ".mlp", ".mov", ".mp4", ".m4a", ".3gp", ".3g2", ".mj2", ".mp2", ".mp3", ".m2a", ".mpa", ".mpc", ".mjpg", ".txt", ".mpl2", ".sub", ".msf", ".mtaf", ".ul", ".musx", ".mvi", ".mxg", ".v", ".nist", ".sph", ".nsp", ".nut", ".ogg", ".oma", ".omg", ".aa3", ".pjs", ".pvf", ".yuv", ".cif", ".qcif", ".rgb", ".rt", ".rsd", ".rsd", ".rso", ".sw", ".sb", ".smi", ".sami", ".sbc", ".msbc", ".sbg", ".scc", ".sdr2", ".sds", ".sdx", ".ser", ".shn", ".vb", ".son", ".sln", ".mjpg", ".stl", ".sub", ".sub", ".sup", ".svag", ".tak", ".thd", ".tta", ".ans", ".art", ".asc", ".diz", ".ice", ".nfo", ".txt", ".vt", ".ty", ".ty+", ".uw", ".ub", ".v210", ".yuv10", ".vag", ".vc1", ".rcv", ".viv", ".idx", ".vpk", ".txt", ".vqf", ".vql", ".vqe", ".vtt", ".wsd", ".xmv", ".xvag", ".yop", ".y4m"}}}
}

func (t *VideoThumb) GenerateThumb(file io.Reader, name string, url string) (*Thumb, error) {
	var err error
	var img image.Image
	sec := rand.Intn(300) + 300
	cmd := exec.Command("ffmpeg", "-ss", strconv.Itoa(sec), "-i", name, "-vframes", "1", "-f", "singlejpeg", "-", "-y")
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err = cmd.Run(); err != nil || buffer.Len() == 0 {
		fallbackCmd := exec.Command("ffmpeg", "-i", name, "-vf", "thumbnail", "-vframes", "1", "-f", "singlejpeg", "-", "-y")
		fallbackCmd.Stdout = &buffer
		if err = fallbackCmd.Run(); err != nil {
			return nil, err
		}
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

func (t *PDFThumb) GenerateThumb(file io.Reader, name string, downLoadUrl string) (*Thumb, error) {
	var err error
	cmd := exec.Command("gs", "-q", "-dBATCH", "-dNOPAUSE", "-sDEVICE=jpeg", "-dFirstPage=1", "-dLastPage=1", "-sOutputFile=-", name)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err = cmd.Run(); err != nil || buffer.Len() == 0 {
		return nil, err
	}
	img, err := jpeg.Decode(&buffer)
	if err != nil {
		return nil, err
	}
	return &Thumb{
		src: img,
		ext: getFileExt(name),
	}, nil
}

type DocThumb struct {
	*BaseThumb
}

func NewDocThumb() *DocThumb {
	return &DocThumb{&BaseThumb{[]string{".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx"}}}
}

func (t *DocThumb) GenerateThumb(file io.Reader, name string, downLoadUrl string) (*Thumb, error) {
	dir, err := ioutil.TempDir("", "*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir) // clean up

	cmd := exec.Command("soffice", "--convert-to", "jpg", "--outdir", dir, name)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer
	if err = cmd.Run(); err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	imgFile, err := os.Open(filepath.Join(dir, files[0].Name()))
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	img, err := jpeg.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	return &Thumb{
		src: img,
		ext: getFileExt(name),
	}, nil
}
