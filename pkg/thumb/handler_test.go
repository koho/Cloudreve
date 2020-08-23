package thumb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestVideoThumb_GenerateThumb(t *testing.T) {
	asserts := assert.New(t)
	fileName := "/home/gerhard/GolandProjects/Cloudreve/cutout1.mp4"
	vt := NewVideoThumb()
	asserts.True(vt.CanHandle(fileName))
	file, err := os.Open(fileName)
	if asserts.NoError(err) {
		defer file.Close()
		_, err = vt.GenerateThumb(file, fileName)
		asserts.NoError(err)
	}
}
