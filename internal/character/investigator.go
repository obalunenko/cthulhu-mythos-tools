package character

import (
	"encoding/json"
)

func UnmarshalInvestigator(data []byte) (Investigator, error) {
	var r Investigator

	if err := json.Unmarshal(data, &r); err != nil {
		return Investigator{}, err
	}

	return r, nil
}

func (r *Investigator) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Investigator struct {
	Investigator InvestigatorClass `json:"Investigator"`
}

type InvestigatorClass struct {
	Header          Header          `json:"Header"`
	PersonalDetails PersonalDetails `json:"PersonalDetails"`
	Characteristics Characteristics `json:"Characteristics"`
	Skills          Skills          `json:"Skills"`
	Talents         any             `json:"Talents"`
	Weapons         Weapons         `json:"Weapons"`
	Combat          Combat          `json:"Combat"`
	Backstory       Backstory       `json:"Backstory"`
	Possessions     Possessions     `json:"Possessions"`
	Cash            Cash            `json:"Cash"`
	Assets          any             `json:"Assets"`
}

type Backstory struct {
	Description string `json:"description"`
	Traits      string `json:"traits"`
	Ideology    string `json:"ideology"`
	Injurues    any    `json:"injurues"`
	People      string `json:"people"`
	Phobias     string `json:"phobias"`
	Locations   string `json:"locations"`
	Tomes       string `json:"tomes"`
	Possessions string `json:"possessions"`
	Encounters  string `json:"encounters"`
}

type Cash struct {
	Spending string `json:"spending"`
	Cash     string `json:"cash"`
	Assets   string `json:"assets"`
}

type Characteristics struct {
	Str                         string `json:"STR"`
	Dex                         string `json:"DEX"`
	Int                         string `json:"INT"`
	Con                         string `json:"CON"`
	App                         string `json:"APP"`
	Pow                         string `json:"POW"`
	Siz                         string `json:"SIZ"`
	Edu                         string `json:"EDU"`
	Move                        string `json:"Move"`
	Luck                        string `json:"Luck"`
	LuckMax                     string `json:"LuckMax"`
	Sanity                      string `json:"Sanity"`
	SanityStart                 string `json:"SanityStart"`
	SanityMax                   string `json:"SanityMax"`
	MagicPts                    string `json:"MagicPts"`
	MagicPtsMax                 string `json:"MagicPtsMax"`
	HitPts                      string `json:"HitPts"`
	HitPtsMax                   string `json:"HitPtsMax"`
	DamageBonus                 string `json:"DamageBonus"`
	Build                       string `json:"Build"`
	OccupationSkillPoints       string `json:"OccupationSkillPoints"`
	PersonalInterestSkillPoints string `json:"PersonalInterestSkillPoints"`
}

type Combat struct {
	DamageBonus string `json:"DamageBonus"`
	Build       string `json:"Build"`
	Dodge       Dodge  `json:"Dodge"`
}

type Dodge struct {
	SkillValues
}

type Header struct {
	Title       string `json:"Title"`
	Creator     string `json:"Creator"`
	CreateDate  string `json:"CreateDate"`
	GameName    string `json:"GameName"`
	GameVersion string `json:"GameVersion"`
	GameType    string `json:"GameType"`
	Discalimer  string `json:"Discalimer"`
	Version     string `json:"Version"`
}

type PersonalDetails struct {
	Name       string `json:"Name"`
	Occupation string `json:"Occupation"`
	Archetype  any    `json:"Archetype"`
	Gender     string `json:"Gender"`
	Age        string `json:"Age"`
	Birthplace string `json:"Birthplace"`
	Residence  string `json:"Residence"`
	Portrait   string `json:"Portrait"`
}

type Possessions struct {
	Item Item `json:"item"`
}

type Item struct {
	Description string `json:"description"`
}

type Skills struct {
	Skill []Skill `json:"Skill"`
}

type SkillValues struct {
	Value string `json:"value"`
	Half  string `json:"half"`
	Fifth string `json:"fifth"`
}

type Skill struct {
	Name string `json:"name"`
	SkillValues
	Subskill   *string `json:"subskill,omitempty"`
	Occupation *string `json:"occupation,omitempty"`
}

type Weapons struct {
	Weapon []Weapon `json:"weapon"`
}

type Weapon struct {
	Name      string  `json:"name"`
	Skillname string  `json:"skillname"`
	Regular   string  `json:"regular"`
	Hard      *string `json:"hard"`
	Extreme   *string `json:"extreme"`
	Damage    string  `json:"damage"`
	Range     string  `json:"range"`
	Attacks   string  `json:"attacks"`
	Ammo      string  `json:"ammo"`
	Malf      string  `json:"malf"`
}
