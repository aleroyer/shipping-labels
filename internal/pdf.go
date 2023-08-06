package internal

import (
	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rs/zerolog/log"
	"regexp"
	"strconv"
)

func PDFInfo(path string) []string {
	info, err := pdf.InfoFile(path, nil, nil)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil
	}
	return info
}

func PDFPageSize(info []string) (width float64, height float64) {
	re := regexp.MustCompile(`Page size: ([0-9\.]+) x ([0-9\.]+) points`)
	matches := re.FindAllStringSubmatch(info[3], -1)
	for _, match := range matches {
		width, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		height, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		return width, height
	}
	return 0, 0
}

func PDFCrop(in string, out string, box *model.Box) error {
	return pdf.CropFile(in, out, nil, box, nil)
}

func PDFCombine(in []string, out string) error {
	return pdf.MergeAppendFile(in, out, nil)
}

func PDFRotate(in string, out string, rotation int) error {
	return pdf.RotateFile(in, out, rotation, nil, nil)
}
