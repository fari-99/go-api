package generate_pdf

import (
	"os/exec"
)

func (r *RequestPdf) GeneratePDFPuppeteer(htmlPath string, pdfPath string) error {
	cmd := exec.Command("node", "printPDF.js", htmlPath, pdfPath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
