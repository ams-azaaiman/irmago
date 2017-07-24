package irmago

import (
	"encoding/xml"

	"github.com/mhe/gabi"
)

// SchemeManager describes a scheme manager.
type SchemeManager struct {
	ID                string           `xml:"Id"`
	Name              TranslatedString `xml:"Name"`
	URL               string           `xml:"Contact"`
	Description       TranslatedString
	KeyshareServer    string
	KeyshareWebsite   string
	KeyshareAttribute string
	XMLVersion        int      `xml:"version,attr"`
	XMLName           xml.Name `xml:"SchemeManager"`
}

// Issuer describes an issuer.
type Issuer struct {
	ID              string           `xml:"ID"`
	Name            TranslatedString `xml:"Name"`
	ShortName       TranslatedString `xml:"ShortName"`
	SchemeManagerID string           `xml:"SchemeManager"`
	ContactAddress  string
	ContactEMail    string
	URL             string `xml:"baseURL"`
	XMLVersion      int    `xml:"version,attr"`
}

// CredentialType is a description of a credential type, specifying (a.o.) its name, issuer, and attributes.
type CredentialType struct {
	ID              string           `xml:"CredentialID"`
	Name            TranslatedString `xml:"Name"`
	ShortName       TranslatedString `xml:"ShortName"`
	IssuerID        string           `xml:"IssuerID"`
	SchemeManagerID string           `xml:"SchemeManager"`
	IsSingleton     bool             `xml:"ShouldBeSingleton"`
	Description     TranslatedString
	Attributes      []AttributeDescription `xml:"Attributes>Attribute"`
	XMLVersion      int                    `xml:"version,attr"`
	XMLName         xml.Name               `xml:"IssueSpecification"`
}

// AttributeDescription is a description of an attribute within a credential type.
type AttributeDescription struct {
	ID          string `xml:"id,attr"`
	Name        TranslatedString
	Description TranslatedString
}

// TranslatedString represents an XML tag containing a string translated to multiple languages.
// For example: <Foo id="bla"><Translation lang="en">Hello world</Translation><Translation lang="nl">Hallo wereld</Translation></Foo>
// type TranslatedString struct {
// 	Translations []struct {
// 		Language string `xml:"lang,attr"`
// 		Value    string `xml:",chardata"`
// 	} `xml:"Translation"`
// 	ID string `xml:"id,attr"`
// }
//
// // Get returns the specified translation
// func (ts TranslatedString) Get(lang string) string {
// 	for _, l := range ts.Translations {
// 		if l.Language == lang {
// 			return l.Value
// 		}
// 	}
// 	return ""
// }

// TranslatedString represents an XML tag containing a string translated to multiple languages.
// For example: <Foo id="bla"><en>Hello world</en><nl>Hallo wereld</nl></Foo>
type TranslatedString struct {
	Translations []struct {
		XMLName xml.Name
		Text    string `xml:",chardata"`
	} `xml:",any"`
}

// Translation returns the specified translation.
func (ts *TranslatedString) Translation(lang string) string {
	for _, translation := range ts.Translations {
		if translation.XMLName.Local == lang {
			return translation.Text
		}
	}
	return ""
}

// Identifier returns the identifier of the specified credential type.
func (cd *CredentialType) Identifier() string {
	return cd.SchemeManagerID + "." + cd.IssuerID + "." + cd.ID
}

// IssuerIdentifier returns the issuer identifier of the specified credential type.
func (cd *CredentialType) IssuerIdentifier() string {
	return cd.SchemeManagerID + "." + cd.IssuerID
}

// Identifier returns the identifier of the specified issuer description.
func (id *Issuer) Identifier() string {
	return id.SchemeManagerID + "." + id.ID
}

// CurrentPublicKey returns the latest known public key of the issuer identified by this instance.
func (id *Issuer) CurrentPublicKey() *gabi.PublicKey {
	keys := MetaStore.PublicKeys[id.Identifier()]
	if keys == nil || len(keys) == 0 {
		return nil
	}
	return keys[len(keys)-1]
}

// PublicKey returns the specified public key of the issuer identified by this instance.
func (id *Issuer) PublicKey(index int) *gabi.PublicKey {
	keys := MetaStore.PublicKeys[id.Identifier()]
	if keys == nil || index >= len(keys) {
		return nil
	}
	return keys[index]
}
