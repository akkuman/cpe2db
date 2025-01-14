package main

import (
	"github.com/knqyf263/go-cpe/common"
	"github.com/knqyf263/go-cpe/naming"
)

// CPEItem 表示 XML 中的 <cpe-item> 元素
type CPEItem struct {
	Name       string      `xml:"name,attr"`
	Title      string      `xml:"title"`
	References []Reference `xml:"references>reference"`
	CPE23Item  CPE23Item   `xml:"http://scap.nist.gov/schema/cpe-extension/2.3 cpe23-item"`
}

// Reference 表示 XML 中的 <reference> 元素
type Reference struct {
	Href string `xml:"href,attr"`
	Text string `xml:",chardata"`
}

// CPE23Item 表示 XML 中的 <cpe23-item> 元素
type CPE23Item struct {
	Name string `xml:"name,attr"`
}

type CPE23Reference struct {
	Type string `json:"type"`
	Link string `json:"link"`
}

type CPE23 struct {
	CPEVer     string           `gorm:"column:cpe_ver"`
	Title      string           `gorm:"column:title"`
	Part       string           `gorm:"column:part"`
	Vendor     string           `gorm:"column:vendor"`
	Product    string           `gorm:"column:product"`
	Version    string           `gorm:"column:version"`
	Update     string           `gorm:"column:update"`
	Edition    string           `gorm:"column:edition"`
	Language   string           `gorm:"column:language"`
	SwEdition  string           `gorm:"column:sw_edition"`
	TargetSw   string           `gorm:"column:target_sw"`
	TargetHw   string           `gorm:"column:target_hw"`
	Other      string           `gorm:"column:other"`
	References []CPE23Reference `gorm:"column:references;serializer:json"` // 存储 JSON 字符串
}

func (CPE23) TableName() string {
	return "cpe23"
}

func ParseCPE23(item CPEItem) (*CPE23, error) {
	wfn, err := naming.UnbindFS(item.CPE23Item.Name)
	if err != nil {
		return nil, err
	}
	cpe23 := &CPE23{
		CPEVer:    "2.3",
		Title:     item.Title,
		Part:      wfn.GetString(common.AttributePart),
		Vendor:    wfn.GetString(common.AttributeVendor),
		Product:   wfn.GetString(common.AttributeProduct),
		Version:   wfn.GetString(common.AttributeVersion),
		Update:    wfn.GetString(common.AttributeUpdate),
		Edition:   wfn.GetString(common.AttributeEdition),
		Language:  wfn.GetString(common.AttributeLanguage),
		SwEdition: wfn.GetString(common.AttributeSwEdition),
		TargetSw:  wfn.GetString(common.AttributeTargetSw),
		TargetHw:  wfn.GetString(common.AttributeTargetHw),
		Other:     wfn.GetString(common.AttributeOther),
	}
	for _, v := range item.References {
		cpe23.References = append(cpe23.References, CPE23Reference{
			Type: v.Text,
			Link: v.Href,
		})
	}
	return cpe23, nil
}
