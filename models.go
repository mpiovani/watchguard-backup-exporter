package main

import (
	"encoding/xml"
)

type LoginResponse struct {
	XMLName xml.Name `xml:"methodResponse"`
	Text    string   `xml:",chardata"`
	Params  struct {
		Text  string `xml:",chardata"`
		Param struct {
			Text  string `xml:",chardata"`
			Value struct {
				Text   string `xml:",chardata"`
				Struct struct {
					Text   string `xml:",chardata"`
					Member []struct {
						Text  string `xml:",chardata"`
						Name  string `xml:"name"`
						Value struct {
							Text   string `xml:",chardata"`
							Struct struct {
								Text   string `xml:",chardata"`
								Member []struct {
									Text  string `xml:",chardata"`
									Name  string `xml:"name"`
									Value string `xml:"value"`
								} `xml:"member"`
							} `xml:"struct"`
						} `xml:"value"`
					} `xml:"member"`
				} `xml:"struct"`
			} `xml:"value"`
		} `xml:"param"`
	} `xml:"params"`
}

type FileActionResponse struct {
	XMLName xml.Name `xml:"methodResponse"`
	Text    string   `xml:",chardata"`
	Params  struct {
		Text  string `xml:",chardata"`
		Param struct {
			Text  string `xml:",chardata"`
			Value struct {
				Text   string `xml:",chardata"`
				String string `xml:"string"`
			} `xml:"value"`
		} `xml:"param"`
	} `xml:"params"`
}
