package tor

import "strings"

var Eyes5C []string = []string{
	"{US}", "{GB}", "{AU}", "{CA}", "{NZ}"}

var Eyes9C []string = []string{
	"{US}", "{GB}", "{AU}", "{CA}", "{NZ}", "{FR}", "{GB}", "{AU}", "{CA}", "{NZ}"}

var Eyes14C []string = []string{
	"{US}", "{GB}", "{AU}", "{CA}", "{NZ}", "{FR}", "{GB}", "{AU}", "{CA}", "{NZ}", "{BE}", "{IT}", "{ES}", "{DE}", "{SE}"}

var Eyes14CPlus []string = []string{
	"{US}", "{GB}", "{AU}", "{CA}", "{NZ}", "{FR}", "{GB}", "{AU}", "{CA}", "{NZ}", "{BE}", "{IT}", "{ES}", "{DE}", "{SE}", "{KP}", "{IL}", "{CN}", "{KR}", "{IL}", "{SG}"}

const Countries string = "{AF},{AX},{AL},{DZ},{AD},{AO},{AI},{AQ},{AG},{AR},{AM},{AW},{AU},{AT},{AZ},{BS},{BH},{BD},{BB},{BY},{BE},{BZ},{BJ},{BM},{BT},{BO},{BA},{BW},{BV},{BR},{IO},{VG},{BN},{BG},{BF},{BI},{KH},{CM}, {CA},{CV},{KY},{CF},{TD},{CL},{CN},{CX},{CC},{CO},{KM},{CG},{CD},{CK},{CR},{CI},{HR},{CU},{CY},{CZ},{DK},{DJ},{DM},{DO},{EC},{EG},{SV},{GQ},{EE},{ET},{FK},{FO},{FJ},{FI},{FR},{GF},{PF},{TF},{GA},{GM},{GE},{DE},{GH},{GI},{GR},{GL},{GD},{GP},{GU},{GT},{GN},{GW},{GY},{HT},{HM},{HN},{HK},{HU},{IS},{IN},{ID},{IR},{IQ},{IE},{IM},{IL},{IT},{JM},{JP},{JO},{KZ},{KE},{KI},{KP},{KR},{KW},{KG},{LA},{LV},{LB},{LS},{LR},{LY},{LI},{LT},{LU},{MO},{MK},{MG},{MW},{MY},{MV},{ML},{MT},{MH},{MQ},{MR},{MU},{YT},{MX},{FM},{MD},{MC},{MN},{ME},{MS},{MA},{MZ},{MM},{NA},{NR},{NP},{NL},{NC},{NZ},{NI},{NE},{NG},{NU},{NF},{MP},{NO},{OM},{PK},{PW},{PS},{PA},{PG},{PY},{PE},{PH},{PN},{PL},{PT},{PR},{QA},{RE},{RO},{RU},{RW},{WS},{SM},{ST},{SA},{SN},{RS},{SC},{SL},{SG},{SK},{SI},{SB},{SO},{AS},{ZA},{GS},{ES},{LK},{SH},{KN},{LC},{PM},{VC},{SD},{SR},{SJ},{SZ},{SE},{CH},{SY},{TW},{TJ},{TZ},{TH},{TG},{TK},{TO},{TT},{TN},{TR},{TM},{TC},{TV},{UG},{UA},{AE},{GB},{US},{UM},{UY},{UZ},{VU},{VA},{VE},{VN},{VI},{WF},{EH},{YE},{ZM},{ZW}"

var excludedTorAddress []string = []string{
	"0.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"255.255.255.255/32",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"192.88.99.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
}

func NonTor() string {
	return strings.Join(excludedTorAddress, " ")
}

func NoEyes(list []string) string {
	var c = Countries
	for _, r := range list {
		c = strings.Replace(c, r+",", "", 1)
	}
	return c
}
