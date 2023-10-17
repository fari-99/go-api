package generate_pdf

import (
	"fmt"
	"log"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

//generate pdf function
func (r *RequestPdf) GeneratePDFWkHtmlToPdf(pdfPath string) (bool, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return false, fmt.Errorf("error get base pdf generator, err := %s", err.Error())
	}

	pageReader := wkhtmltopdf.NewPageReader(strings.NewReader(r.body))
	pageReader.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(pageReader)

	err = pdfg.Create()
	if err != nil {
		return false, fmt.Errorf("error create pdf fileLocation buffer, err := %s", err.Error())
	}

	err = pdfg.WriteFile(pdfPath)
	if err != nil {
		return false, fmt.Errorf("error create pdf fileLocation, err := %s", err.Error())
	}

	log.Printf("success create pdf files")
	return true, nil
}
