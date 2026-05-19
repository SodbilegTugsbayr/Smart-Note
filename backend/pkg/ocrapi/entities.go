package ocrapi

type OCRResult struct {
	FileName  string `json:"file_name"`
	PageCount int    `json:"page_count"`
	UUID      string `json:"uuid"`
}

type OCRTranscript struct {
	Done   bool `json:"done"`
	Result struct {
		FailedPages []int  `json:"failed_pages"`
		File        string `json:"file"`
		OCR         []struct {
			Page int    `json:"page"`
			Text string `json:"text"`
		} `json:"ocr"`
		SuccessPages []int `json:"success_pages"`
	} `json:"result"`
}
