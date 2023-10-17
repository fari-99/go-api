package generate_pdf

import (
	"fmt"
	"os"
	"testing"
)

func TestCreatePDFCss(t *testing.T) {
	basePDF := NewRequestPdf("")
	templateLocation := "./generate_pdf/"
	err := basePDF.ParseTemplate(templateLocation+"test_css.html", nil)
	if err != nil {
		t.Error(fmt.Sprintf("error parse template, err := %s", err.Error()))
		t.Fail()
		return
	}

	filename := "test_local"
	fileType := "pdf"

	//tempLocation := "../storages/"
	tempLocation := "../"
	tempPdfPath := fmt.Sprintf("%s%s.%s", tempLocation, filename, fileType)
	if ok, err := basePDF.GeneratePDFWkHtmlToPdf(tempPdfPath); err != nil {
		t.Error(fmt.Sprintf("error generate pdf, err := %s", err.Error()))
		t.Fail()
		return
	} else if !ok {
		t.Error("failed to generate pdf")
		t.Fail()
		return
	}
	//defer os.Remove(tempPdfPath) // delete files after done [delete if you want to check file that got generated]
	return
}

func TestCreatePDF(t *testing.T) {
	basePDF := NewRequestPdf("")
	templateLocation := "./generate_pdf/"
	err := basePDF.ParseTemplate(templateLocation+"test.html", nil)
	if err != nil {
		t.Error(fmt.Sprintf("error parse template, err := %s", err.Error()))
		t.Fail()
		return
	}

	filename := "test_internet"
	fileType := "pdf"

	//tempLocation := "../storages/"
	tempLocation := "../"
	tempPdfPath := fmt.Sprintf("%s%s.%s", tempLocation, filename, fileType)
	if ok, err := basePDF.GeneratePDFWkHtmlToPdf(tempPdfPath); err != nil {
		t.Error(fmt.Sprintf("error generate pdf, err := %s", err.Error()))
		t.Fail()
		return
	} else if !ok {
		t.Error("failed to generate pdf")
		t.Fail()
		return
	}
	defer os.Remove(tempPdfPath) // delete files after done [delete if you want to check file that got generated]
	return
}
