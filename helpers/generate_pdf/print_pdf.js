const puppeteer = require('puppeteer');

async function printPDF(url, filename) {
    const browser = await puppeteer.launch({headless:true});
    const page = await browser.newPage();
    await page.goto('file://'+url, {waitUntil:'networkidle0'});
    const pdf = await page.pdf({
        format: 'A4',
        margin: {
            top: "20px",
            left: "20px",
            right: "20px",
            bottom: "20px"
        }
    });

    await browser.close();
    return pdf
}

const url = process.argv[2];
const filename = process.argv[3];

printPDF(url, filename).then(pdf => {
    require('fs').writeFileSync(filename, pdf);
});