package spider

import "gorm.io/gorm"

type Patent struct {
	gorm.Model

	Title           string // 标题
	Url             string // 专利的url
	NaviCode        string // 学科代码
	Year            string // 年份, 应该是公开日的年份，仅作爬虫分类用，不一定准确
	ApplicationType string // 专利类型
	ApplicationDate string // 申请日
	//PublicationNo        string `gorm:"index:idx_public_no,unique"`       // 申请公布号/授权公布号
	ApplyPublicationNo   string `gorm:"index:idx_apply_public_no,unique"` // 申请公布号
	AuthPublicationNo    string `gorm:"index:idx_auth_public_no,unique"`  // 授权公布号
	MultiPublicationNo   string // 多次公布
	PublicationDate      string // 公开公告日
	AuthPublicationDate  string //授权公告日
	Applicant            string // 申请人
	ApplicantAddress     string // 地址
	Inventors            string // 发明人原始字符串
	ApplicationNO        string // 申请(专利)号
	AreaCode             string // 国省代码
	ClassificationNO     string // 分类号
	MainClassificationNo string // 主分类号
	Agency               string // 代理机构
	Agent                string // 代理人
	Page                 string // 页数
	Abstract             string // 摘要
	Sovereignty          string // 主权项
	LegalStatus          string // 法律状态
}

// FillRowFields 填充专利的字段
func (patent *Patent) FillRowFields(key, value string) {
	switch key {
	case "专利类型：":
		patent.ApplicationType = value
	case "申请日：":
		patent.ApplicationDate = value
	// 多次公布后面发现是动态加载，其实这样获取不到，但是，不同阶段的专利的申请（专利）号是一样的，多次公布不爬问题也不大
	case "多次公布：":
		patent.MultiPublicationNo = value
	case "申请人：":
		patent.Applicant = value
	case "地址：":
		patent.ApplicantAddress = value
	case "发明人：":
		patent.Inventors = value
	case "申请（专利）号：":
		patent.ApplicationNO = value
	case "申请公布号：":
		patent.ApplyPublicationNo = value
	case "授权公布号：":
		patent.AuthPublicationNo = value
	case "公开公告日：":
		patent.PublicationDate = value
	case "授权公告日：":
		patent.AuthPublicationDate = value
	case "国省代码：":
		patent.AreaCode = value
	case "分类号：":
		patent.ClassificationNO = value
	case "主分类号：":
		patent.MainClassificationNo = value
	case "代理机构：":
		patent.Agency = value
	case "代理人":
		patent.Agent = value
	case "页数":
		patent.Page = value
	}
}

// Validate 校验专利的字段，如果空值太多，则失败
func (patent *Patent) Validate() bool {
	if patent.Title == "" {
		return false
	}
	// 至少应用 8 个字段不为空
	notEmptyCount := 0
	for _, field := range []string{
		patent.ApplicationType,
		patent.ApplicationDate,
		patent.ApplyPublicationNo,
		patent.AuthPublicationNo,
		patent.PublicationDate,
		patent.AuthPublicationDate,
		patent.Applicant,
		patent.ApplicantAddress,
		patent.Inventors,
		patent.ApplicationNO,
		patent.AreaCode,
		patent.ClassificationNO,
		patent.MainClassificationNo,
		patent.Agency,
		patent.Agent,
		patent.Page,
	} {
		if field != "" {
			notEmptyCount++
		}
	}
	return notEmptyCount >= 8
}
