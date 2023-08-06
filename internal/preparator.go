package internal

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpuModel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"os"
	"path"
	"strings"
	"time"
)

type Preparator struct {
	fs      afero.Fs
	srcDir  string
	destDir string
}

func NewPreparator(src string, dest string) (p *Preparator, err error) {
	p = &Preparator{}
	p.fs = afero.NewOsFs()

	if p.validateDirectory(src) {
		log.Info().Msgf("Setting source directory to: %s", src)
		p.srcDir = src
	} else {
		return nil, fmt.Errorf("source directory %s is invalid", src)
	}

	if p.validateDirectory(dest) {
		log.Info().Msgf("Setting destination directory to: %s", dest)
		p.destDir = dest
	} else {
		return nil, fmt.Errorf("destination directory %s is invalid", dest)
	}

	return p, nil
}

func (p *Preparator) Prepare() (err error) {
	log.Info().Msg("Starting preparation of shipping labels")
	srcDirContent, err := afero.ReadDir(p.fs, p.srcDir)
	if err != nil {
		return err
	}
	pdfFiles := p.filterPDFFiles(srcDirContent)
	var filesToMerge []string
	for _, pdf := range pdfFiles {
		pdfInfo := PDFInfo(path.Join(p.srcDir, pdf.Name()))
		box, err := p.getCropbox(pdfInfo)
		if err != nil {
			return err
		}
		croppedFileName := path.Join(p.destDir, fmt.Sprintf(".cropped.%s", pdf.Name()))
		log.Info().Msgf("Cropping file %s to %s", pdf.Name(), croppedFileName)
		err = PDFCrop(path.Join(p.srcDir, pdf.Name()), croppedFileName, box)
		if err != nil {
			return err
		}
		if p.getProvider(pdfInfo) == "Mondial Relay" {
			log.Info().Msg("Got a Mondial Relay PDF, rotating result")
			err = PDFRotate(croppedFileName, "", 270)
			if err != nil {
				return err
			}
		}
		filesToMerge = append(filesToMerge, croppedFileName)
	}
	log.Info().Msg("All files cropped!")

	log.Info().Msg("Combining cropped files into one...")
	dateStr := time.Now().Format("20060102-1504")
	outputFileName := path.Join(p.destDir, fmt.Sprintf("%s_%s.pdf", dateStr, "ready_to_print"))
	err = PDFCombine(filesToMerge, outputFileName)
	if err != nil {
		return err
	}
	log.Info().Msgf("All files combined to %s", outputFileName)

	log.Info().Msg("Cleaning up...")
	err = p.cleanup(filesToMerge)
	if err != nil {
		return err
	}

	return nil
}

func (p *Preparator) cleanup(files []string) error {
	for _, f := range files {
		err := p.fs.Remove(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Preparator) getCropbox(pdfInfo []string) (b *pdfcpuModel.Box, err error) {
	pageWidth, pageHeight := PDFPageSize(pdfInfo)
	provider := p.getProvider(pdfInfo)
	b = &pdfcpuModel.Box{}
	switch provider {
	case "Mondial Relay":
		log.Info().Msg("Provider identified as Mondial Relay")
		// Left, Bottom, Right, Top
		b, err = api.Box(fmt.Sprintf("[%.2f %.2f %.2f %.2f]", 0.0, pageHeight/2, pageWidth, pageHeight), types.POINTS)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	case "Colissimo":
		log.Info().Msg("Provider identified as Colissimo")
		// Left, Bottom, Right, Top
		b, err = api.Box(fmt.Sprintf("[%.2f %.2f %.2f %.2f]", 52.0, 160.0, (pageWidth/2)-76, pageHeight-85), types.POINTS)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	default:
		return b, fmt.Errorf("no cropbox defined for provider %s", provider)
	}

	return b, nil
}

func (p *Preparator) getProvider(pdfInfo []string) string {
	if strings.Contains(pdfInfo[6], "Author: MondialRelay") {
		return "Mondial Relay"
	}
	if strings.HasSuffix(pdfInfo[6], "Author: ") && strings.Contains(pdfInfo[8], "PDF Producer: iText") {
		return "Colissimo"
	}
	return ""
}

func (p *Preparator) filterPDFFiles(files []os.FileInfo) (filtered []os.FileInfo) {
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".pdf") {
			log.Debug().Msgf("%s extension is not .pdf, ignoring", file.Name())
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func (p *Preparator) validateDirectory(path string) (valid bool) {
	dir, err := afero.DirExists(p.fs, path)
	if err != nil {
		log.Error().Msgf("Directory %s doesn't exist", path)
		return
	}
	if dir {
		isDir, err := afero.IsDir(p.fs, path)
		if err != nil {
			log.Error().Msgf("%s is not a directory", path)
			return
		}
		return isDir
	}
	return
}
