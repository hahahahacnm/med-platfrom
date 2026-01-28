const XLSX = require('xlsx');
const path = require('path');

// Allow passing filename as argument
const fileName = '1.绪论_2026_1_14.xlsx';
const filePath = path.join(__dirname, '../' + fileName);

try {
    const workbook = XLSX.readFile(filePath);
    const sheetName = workbook.SheetNames[0];
    const sheet = workbook.Sheets[sheetName];
    const data = XLSX.utils.sheet_to_json(sheet);

    console.log(`Total rows: ${data.length}`);

    let nonChoiceCount = 0;

    for (let i = 0; i < data.length; i++) {
        const row = data[i];
        if (!row['题干']) continue;

        const hasOption = row['选项A'] || row['选项B'] || row['选项C'] || row['选项D'] || row['选项E'];
        if (!hasOption) {
            nonChoiceCount++;
            if (nonChoiceCount <= 3) {
                console.log('--- Non-Choice Question Found ---');
                console.log('Question:', row['题干']);
                console.log('Answer:', row['正确答案']);
                console.log('Explanation:', row['解析']);
            }
        }
    }
    console.log(`Found ${nonChoiceCount} non-choice questions.`);

} catch (error) {
    console.error(error);
}
