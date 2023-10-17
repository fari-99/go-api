package generate_pdf

import "testing"

func TestPuppeteerGeneratePDF(t *testing.T) {
	htmlPath := "/home/fari_99/my_project/main-project/workspace/go-api/helpers/generate_pdf/template/test_css.html"
	outputPath := "./output_puppeteer.pdf"

	basePDF := NewRequestPdf("")
	err := basePDF.GeneratePDFPuppeteer(htmlPath, outputPath)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
		return
	}
}
