package aws

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v3"

	"github.com/linweiyuan/aws-api/internal/proxy"
)

const (
	loginUrl = "" // aws login url
)

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type XmlResponse struct {
	Assertion Assertion
}

type Assertion struct {
	AttributeStatement AttributeStatement
}

type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

type Attribute struct {
	Values []string `xml:"AttributeValue"`
}

type LoginResponse struct {
	RoleInfo      []RoleInfo `json:"roleInfo"`
	SamlAssertion string     `json:"samlAssertion"`
}

type RoleInfo struct {
	PrincipalArn string `json:"principalArn"`
	RoleArn      string `json:"roleArn"`
}

func Login(c fiber.Ctx) error {
	userInfo := new(UserInfo)
	if err := c.Bind().JSON(userInfo); err != nil {
		return err
	}

	response := userInfo.LoginAWS()
	if response == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to login aws"})
	}

	return c.JSON(response)
}

func (userInfo UserInfo) LoginAWS() *LoginResponse {
	urlParams := url.Values{
		"Username": {userInfo.Username},
		"Password": {userInfo.Password},
	}

	req, _ := http.NewRequest(http.MethodPost, loginUrl, strings.NewReader(urlParams.Encode()))
	resp, _ := proxy.NewClient().Do(req)
	defer resp.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	samlAssertion, _ := doc.Find("input[name=SAMLResponse]").Attr("value")
	decodedXml, _ := base64.StdEncoding.DecodeString(samlAssertion)

	var xmlResponse XmlResponse
	err := xml.Unmarshal(decodedXml, &xmlResponse)
	if err != nil {
		return nil
	}

	if len(xmlResponse.Assertion.AttributeStatement.Attributes) == 0 {
		return nil
	}

	var roleInfoList []RoleInfo
	for _, attributeValue := range xmlResponse.Assertion.AttributeStatement.Attributes[0].Values {
		parts := strings.Split(attributeValue, ",")
		roleInfoList = append(roleInfoList, RoleInfo{
			PrincipalArn: parts[0],
			RoleArn:      parts[1],
		})
	}

	return &LoginResponse{
		RoleInfo:      roleInfoList,
		SamlAssertion: samlAssertion,
	}
}
