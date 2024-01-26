package tags

import (
	"slices"
	"sort"

	"github.com/jeffyfung/flight-info-agg/pkg/languages"
)

type (
	DestWithLabel struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}

	AirlinesWithLabel struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
)

var Destinations = []map[languages.Lang]string{
	{languages.EN: "Taiwan", languages.TC: "台灣"},
	{languages.EN: "Japan", languages.TC: "日本"},
	{languages.EN: "Korea", languages.TC: "韓國"},
	{languages.EN: "Thailand", languages.TC: "泰國"},
	{languages.EN: "Singapore", languages.TC: "新加坡"},
	{languages.EN: "Malaysia", languages.TC: "馬來西亞"},
	{languages.EN: "Vietnam", languages.TC: "越南"},
	{languages.EN: "Philippines", languages.TC: "菲律賓"},
	{languages.EN: "Indonesia", languages.TC: "印尼"},
	{languages.EN: "Cambodia", languages.TC: "柬埔寨"},
	{languages.EN: "Laos", languages.TC: "老撾"},
	{languages.EN: "Myanmar", languages.TC: "緬甸"},
	{languages.EN: "India", languages.TC: "印度"},
	{languages.EN: "Nepal", languages.TC: "尼泊爾"},
	{languages.EN: "Sri Lanka", languages.TC: "斯里蘭卡"},
	{languages.EN: "China", languages.TC: "中國"},
	{languages.EN: "Macau", languages.TC: "澳門"},
	{languages.EN: "Mongolia", languages.TC: "蒙古"},
	{languages.EN: "Russia", languages.TC: "俄羅斯"},
	{languages.EN: "Australia", languages.TC: "澳洲"},
	{languages.EN: "New Zealand", languages.TC: "紐西蘭"},
	{languages.EN: "United States", languages.TC: "美國"},
	{languages.EN: "Canada", languages.TC: "加拿大"},
	{languages.EN: "United Kingdom", languages.TC: "英國"},
	{languages.EN: "France", languages.TC: "法國"},
	{languages.EN: "Germany", languages.TC: "德國"},
	{languages.EN: "Italy", languages.TC: "意大利"},
	{languages.EN: "Spain", languages.TC: "西班牙"},
	{languages.EN: "Portugal", languages.TC: "葡萄牙"},
	{languages.EN: "Netherlands", languages.TC: "荷蘭"},
	{languages.EN: "Belgium", languages.TC: "比利時"},
	{languages.EN: "Switzerland", languages.TC: "瑞士"},
	{languages.EN: "Austria", languages.TC: "奧地利"},
	{languages.EN: "Czech Republic", languages.TC: "捷克"},
	{languages.EN: "Poland", languages.TC: "波蘭"},
	{languages.EN: "Hungary", languages.TC: "匈牙利"},
	{languages.EN: "Greece", languages.TC: "希臘"},
	{languages.EN: "Turkey", languages.TC: "土耳其"},
	{languages.EN: "Egypt", languages.TC: "埃及"},
	{languages.EN: "South Africa", languages.TC: "南非"},
	{languages.EN: "Brazil", languages.TC: "巴西"},
	{languages.EN: "Argentina", languages.TC: "阿根廷"},
	{languages.EN: "Chile", languages.TC: "智利"},
	{languages.EN: "Mexico", languages.TC: "墨西哥"},
	{languages.EN: "Peru", languages.TC: "秘魯"},
	{languages.EN: "Colombia", languages.TC: "哥倫比亞"},
	{languages.EN: "Ecuador", languages.TC: "厄瓜多爾"},
	{languages.EN: "Panama", languages.TC: "巴拿馬"},
	{languages.EN: "Costa Rica", languages.TC: "哥斯達黎加"},
	{languages.EN: "Cuba", languages.TC: "古巴"},
	{languages.EN: "Dominican Republic", languages.TC: "多米尼加"},
	{languages.EN: "Puerto Rico", languages.TC: "波多黎各"},
	{languages.EN: "Jamaica", languages.TC: "牙買加"},
	{languages.EN: "Trinidad and Tobago", languages.TC: "千里達及托巴哥"},
	{languages.EN: "Barbados", languages.TC: "巴貝多"},
	{languages.EN: "Bahamas", languages.TC: "巴哈馬"},
	{languages.EN: "Bermuda", languages.TC: "百慕達"},
	{languages.EN: "Fiji", languages.TC: "斐濟"},
	{languages.EN: "Iceland", languages.TC: "冰島"},
	{languages.EN: "Jordan", languages.TC: "約旦"},
	{languages.EN: "Maldives", languages.TC: "馬爾代夫"},
	{languages.EN: "Papua New Guinea", languages.TC: "巴布亞新畿內亞"},
	{languages.EN: "Azerbaijan", languages.TC: "阿塞拜疆"},
	{languages.EN: "Armenia", languages.TC: "亞美尼亞"},
	{languages.EN: "Kazakhstan", languages.TC: "哈薩克"},
	{languages.EN: "Uzbekistan", languages.TC: "烏茲別克斯坦"},
	{languages.EN: "Dubai", languages.TC: "杜拜"},

	// { languages.EN: "Cayman Islands", languages.TC: "開曼群島" },
	// { languages.EN: "Antigua and Barbuda", languages.TC: "安提瓜及巴布達" },
	// { languages.EN: "Saint Lucia", languages.TC: "聖露西亞" },
	// { languages.EN: "Saint Vincent and the Grenadines", languages.TC: "聖文森及格瑞那丁" },
	// { languages.EN: "Saint Kitts and Nevis", languages.TC: "聖克里斯多福及尼維斯" },
	// { languages.EN: "Aruba", languages.TC: "阿魯巴" },

	// 阿塞拜疆
	// 高加索

}

var AliasToDestMap = map[string][]string{
	"澳紐":  {"澳洲", "紐西蘭"},
	"美加":  {"美國", "加拿大"},
	"星馬":  {"新加坡", "馬來西亞"},
	"泰柬":  {"泰國", "柬埔寨"},
	"越柬":  {"越南", "柬埔寨"},
	"越泰":  {"越南", "泰國"},
	"東南亞": {"新加坡", "馬來西亞", "泰國", "柬埔寨", "越南", "印尼", "老撾", "緬甸", "菲律賓", "汶萊"},
	"新西蘭": {"紐西蘭"},

	"奧克蘭":   {"紐西蘭"},
	"惠靈頓":   {"紐西蘭"},
	"基督城":   {"紐西蘭"},
	"北京":    {"中國"},
	"台中":    {"台灣"},
	"台北":    {"台灣"},
	"台南":    {"台灣"},
	"峴港":    {"越南"},
	"胡志明市":  {"越南"},
	"芽莊":    {"越南"},
	"河內":    {"越南"},
	"首爾":    {"韓國"},
	"釜山":    {"韓國"},
	"濟州":    {"韓國"},
	"東京":    {"日本"},
	"大阪":    {"日本"},
	"名古屋":   {"日本"},
	"福岡":    {"日本"},
	"高松":    {"日本"},
	"鹿兒島":   {"日本"},
	"熊本":    {"日本"},
	"沖繩":    {"日本"},
	"札幌":    {"日本"},
	"仙台":    {"日本"},
	"廣島":    {"日本"},
	"北海道":   {"日本"},
	"布吉":    {"泰國"},
	"曼谷":    {"泰國"},
	"清邁":    {"泰國"},
	"吉隆坡":   {"馬來西亞"},
	"檳城":    {"馬來西亞"},
	"馬尼拉":   {"菲律賓"},
	"宿霧":    {"菲律賓"},
	"長灘島":   {"菲律賓"},
	"雅加達":   {"印尼"},
	"巴里":    {"印尼"},
	"峇里島":   {"印尼"},
	"珀斯":    {"澳洲"},
	"雪梨":    {"澳洲"},
	"悉尼":    {"澳洲"},
	"巴塞隆拿":  {"西班牙"},
	"馬德里":   {"西班牙"},
	"巴黎":    {"法國"},
	"倫敦":    {"英國"},
	"曼徹斯特":  {"英國"},
	"愛丁堡":   {"英國"},
	"都柏林":   {"愛爾蘭"},
	"羅馬":    {"意大利"},
	"米蘭":    {"意大利"},
	"佛羅倫斯":  {"意大利"},
	"威尼斯":   {"意大利"},
	"維也納":   {"奧地利"},
	"薩爾茨堡":  {"奧地利"},
	"慕尼黑":   {"德國"},
	"柏林":    {"德國"},
	"漢堡":    {"德國"},
	"法蘭克福":  {"德國"},
	"科隆":    {"德國"},
	"杜塞爾多夫": {"德國"},
	"斯圖加特":  {"德國"},
	"漢諾威":   {"德國"},
	"紐倫堡":   {"德國"},
	"布拉格":   {"捷克"},
	"布達佩斯":  {"匈牙利"},
	"華沙":    {"波蘭"},
	"克拉科夫":  {"波蘭"},
	"雅典":    {"希臘"},
	"伊斯坦堡":  {"土耳其"},
	"里斯本":   {"葡萄牙"},
	"馬德拉":   {"葡萄牙"},
	"波爾圖":   {"葡萄牙"},
	"布魯塞爾":  {"比利時"},
	"日內瓦":   {"瑞士"},
	"蘇黎世":   {"瑞士"},
	"巴塞爾":   {"瑞士"},
	"阿姆斯特丹": {"荷蘭"},
	"鹿特丹":   {"荷蘭"},
	"馬斯垂克":  {"荷蘭"},
	"奧斯陸":   {"挪威"},
	"斯德哥爾摩": {"瑞典"},
	"赫爾辛基":  {"芬蘭"},
	"雷克雅維克": {"冰島"},
	"開羅":    {"埃及"},
	"開普敦":   {"南非"},
	"約翰內斯堡": {"南非"},
	"墨爾本":   {"澳洲"},
	"布里斯班":  {"澳洲"},
	"阿德雷德":  {"澳洲"},
	"達爾文":   {"澳洲"},
	"堪培拉":   {"澳洲"},
	"霍巴特":   {"澳洲"},
	"紐約":    {"美國"},
	"洛杉磯":   {"美國"},

	"三藩市":   {"美國"},
	"芝加哥":   {"美國"},
	"西雅圖":   {"美國"},
	"波士頓":   {"美國"},
	"華盛頓":   {"美國"},
	"奧蘭多":   {"美國"},
	"邁阿密":   {"美國"},
	"拉斯維加斯": {"美國"},
	"夏威夷":   {"美國"},
	"多倫多":   {"加拿大"},
	"溫哥華":   {"加拿大"},
	"蒙特婁":   {"加拿大"},
	"卡爾加里":  {"加拿大"},
	"埃德蒙頓":  {"加拿大"},
	"渥太華":   {"加拿大"},
	"魁北克":   {"加拿大"},
	"溫尼伯":   {"加拿大"},
	"維多利亞":  {"加拿大"},
	"莫爾斯比港": {"巴布亞新畿內亞"},
}

func DestinationsWithLabels() []DestWithLabel {
	sort.Slice(Destinations, func(i, j int) bool {
		return Destinations[i][languages.EN] < Destinations[j][languages.EN]
	})

	var destList = []DestWithLabel{}
	for _, dest := range Destinations {
		item := dest[languages.TC] + " " + dest[languages.EN]
		destList = append(destList, DestWithLabel{Label: item, Value: dest[languages.TC]})
	}
	return destList
}

var Airlines = []map[languages.Lang]string{
	{languages.EN: "AirAsia", languages.TC: "亞洲航空"},
	{languages.EN: "Cathay Pacific", languages.TC: "國泰航空"},
	{languages.EN: "EVA Air", languages.TC: "長榮航空"},
	{languages.EN: "Hong Kong Airlines", languages.TC: "香港航空"},
	{languages.EN: "Batik Air", languages.TC: "巴澤航空"},
	{languages.EN: "Cebu Pacific", languages.TC: "宿霧太平洋航空"},
	{languages.EN: "China Airlines", languages.TC: "中華航空"},
	{languages.EN: "China Eastern Airlines", languages.TC: "中國東方航空"},
	{languages.EN: "China Southern Airlines", languages.TC: "中國南方航空"},
	{languages.EN: "Garuda Indonesia", languages.TC: "印尼鷹航"},
	{languages.EN: "Jetstar Airways", languages.TC: "捷星航空"},
	{languages.EN: "Jetstar Asia Airways", languages.TC: "捷星亞洲航空"},
	{languages.EN: "Malaysia Airlines", languages.TC: "馬來西亞航空"},
	{languages.EN: "HK Express", languages.TC: "香港快運航空"},

	{languages.EN: "Air France", languages.TC: "法國航空"},
	{languages.EN: "KLM Royal Dutch Airlines", languages.TC: "荷蘭皇家航空"},

	{languages.EN: "Peach Aviation", languages.TC: "樂桃航空"},
	{languages.EN: "Malaysia Airlines", languages.TC: "馬來西亞國際航空"},
	{languages.EN: "Qantas", languages.TC: "澳洲航空"},
	{languages.EN: "Jeju Air", languages.TC: "濟州航空"},
	{languages.EN: "Korean Air", languages.TC: "大韓航空"},
	{languages.EN: "United Airlines", languages.TC: "聯合航空"},
	{languages.EN: "Emirates", languages.TC: "阿聯酋航空"},
	{languages.EN: "Cathay Pacific", languages.TC: "國泰航空"},
	{languages.EN: "Riyadh Air", languages.TC: "利雅得航空"},
	{languages.EN: "Qatar Airways", languages.TC: "卡塔爾航空"},
	{languages.EN: "Etihad Airways", languages.TC: "阿提哈德航空"},
	{languages.EN: "Scoot", languages.TC: "酷航"},
	{languages.EN: "Thai Airways", languages.TC: "泰國航空"},
	{languages.EN: "Cathay Dragon", languages.TC: "港龍航空"},
	{languages.EN: "Asiana Airlines", languages.TC: "韓亞航空"},
	{languages.EN: "Air Canada", languages.TC: "加拿大航空"},

	{languages.EN: "Jetstar Japan", languages.TC: "捷星日本航空"},
	{languages.EN: "Starlux Airlines", languages.TC: "星宇航空"},
	{languages.EN: "Royal Air Philippines", languages.TC: "菲律賓皇家航空"},
	{languages.EN: "Vietjet Air", languages.TC: "越捷航空"},

	{languages.EN: "All Nippon Airways", languages.TC: "全日空"},
	{languages.EN: "Thai Cool Airlines", languages.TC: "泰酷航空"},
	{languages.EN: "Royal Jordanian", languages.TC: "皇家約旦航空"},
	{languages.EN: "Tigerair Taiwan", languages.TC: "台灣虎航"},
	{languages.EN: "Fiji Airways", languages.TC: "斐濟航空"},
	{languages.EN: "T'way Air", languages.TC: "德威航空"},
	{languages.EN: "Vistara", languages.TC: "Vistara"},
	{languages.EN: "Greater Bay Airlines", languages.TC: "大灣區航空"},
	{languages.EN: "Japan Airlines", languages.TC: "日本航空"},
	{languages.EN: "Finnair", languages.TC: "芬蘭航空"},
	{languages.EN: "Air New Zealand", languages.TC: "新西蘭航空"},
	{languages.EN: "Bangkok Airways", languages.TC: "曼谷航空"},
	{languages.EN: "Thai Lion Air", languages.TC: "泰國獅子航空"},
	{languages.EN: "Turkish Airlines", languages.TC: "士耳其航空"},
	{languages.EN: "Air Niugini", languages.TC: "新畿內亞航空"},
	{languages.EN: "Air Macau", languages.TC: "澳門航空"},
	{languages.EN: "Air Busan", languages.TC: "釜山航空"},
	{languages.EN: "China Southern Airlines", languages.TC: "南方航空"},
}

var AliasToAirlineMap = map[string][]string{
	"亞航": {"亞洲航空"},
	"國泰": {"國泰航空"},
	"澳航": {"澳洲航空"},
	"長榮": {"長榮航空"},
	"華航": {"中華航空"},
	"中華": {"中華航空"},
	"馬航": {"馬來西亞國際航空"},
	"港航": {"香港航空"},
	"泰航": {"泰國航空"},

	"日航": {"日本航空"},

	"荷航": {"荷蘭皇家航空"},
	"法航": {"法國航空"},
	"英航": {"英國航空"},
	"加航": {"加拿大航空"},

	"澳洲":  {"澳洲航空"},
	"UO":  {"香港快運航空"},
	"ANA": {"全日空"},
}

func AirlinesWithLabels() []AirlinesWithLabel {
	sort.Slice(Airlines, func(i, j int) bool {
		return Airlines[i][languages.EN] < Airlines[j][languages.EN]
	})

	var airlineList = []AirlinesWithLabel{}
	for _, airline := range Airlines {
		item := airline[languages.TC] + " " + airline[languages.EN]
		airlineList = append(airlineList, AirlinesWithLabel{Label: item, Value: airline[languages.TC]})
	}
	return airlineList
}

func EnrichLocationsWithLabels(locations []string) []DestWithLabel {
	destsWithLabels := DestinationsWithLabels()
	enriched := make([]DestWithLabel, 0, len(locations))
	for _, loc := range locations {
		destLabelPos := slices.IndexFunc(destsWithLabels, func(dest DestWithLabel) bool {
			return dest.Value == loc
		})
		enriched = append(enriched, destsWithLabels[destLabelPos])
	}
	return enriched
}

func EnrichAirlinesWithLabels(airlines []string) []AirlinesWithLabel {
	airlinesWithLabels := AirlinesWithLabels()
	enriched := make([]AirlinesWithLabel, 0, len(airlines))
	for _, airline := range airlines {
		airlineLabelPos := slices.IndexFunc(airlinesWithLabels, func(a AirlinesWithLabel) bool {
			return a.Value == airline
		})
		enriched = append(enriched, airlinesWithLabels[airlineLabelPos])
	}
	return enriched
}
