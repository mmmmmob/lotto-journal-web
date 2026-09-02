package localization

import (
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type Language string

const (
	EN Language = "en"
	TH Language = "th"
)

type Dictionary struct {
	WelcomeFirstTime        string
	WelcomeReturning        string
	SubmitConfirm           string
	SubmitInvalid           string
	SubmitMixed             string
	ListHeader              string
	ListEmpty               string
	LanguageSwitched        string
	AddHelp                 string
	NotifyHelp              string
	WinNotification         string
	LossNotification        string
	DrawWinDetail           string
	DrawLoseDetail          string
	DrawWinMessage          string
	DrawJackpotInfo         string
	DrawFootnote            string
	DrawLoseMessage         string
	GenericError            string
	WelcomeGreetingGeneric  string
	WelcomeGreetingPersonal string
	DbMaintenance           string
	OcrConfirmHeader        string
	OcrWarningHeader        string
	OcrConfirmSave          string
	OcrEdit                 string
	OcrCancel               string
	OcrSaved                string
	OcrCancelled            string
	OcrEditInstruction      string
	OcrPromptCorrection     string
	OcrInvalidCorrection    string
	OcrConflictWarning      string
	OcrShowPending          string
	OcrConfirmPending       string
	OcrCancelPending        string
	OcrBack                 string
}

var Dictionaries = map[Language]*Dictionary{
	EN: {
		WelcomeFirstTime:        "Welcome to Lotto Journal! Send me your ticket numbers (e.g., `123456` or `123456x2` / `456 789`). Type `list` to see your registered tickets.\n\n(Type `ไทย` to switch to Thai language)",
		WelcomeReturning:        "Welcome back to Lotto Journal! Send me your ticket numbers to get started, or type `list` to see your registered tickets.",
		SubmitConfirm:           "Successfully recorded ticket(s) ✅:\n%s",
		SubmitInvalid:           "No valid ticket numbers found ❌\nPlease send 3-digit or 6-digit numbers only.\nInvalid numbers: %s",
		SubmitMixed:             "Successfully recorded ticket(s) ✅:\n%s\n\nInvalid numbers (skipped): %s",
		ListHeader:              "Your tickets for the upcoming draw (%s) 📝:",
		ListEmpty:               "You have no tickets registered for this draw.",
		LanguageSwitched:        "Language changed to English 🇺🇸",
		AddHelp:                 "🎟️ To record ticket numbers, simply send them in chat.\n\n- Single number: `123456` or `456`\n- Multiple numbers: `123456, 654321` or `123 456`\n- Quantity: Append `xN` (e.g. `123456x2` for 2 tickets)",
		NotifyHelp:              "🔔 You will be automatically notified here shortly after the lottery draw finishes (typically around 16:00 Bangkok time on the 1st and 16th of each month).",
		WinNotification:         "Congratulations! Your ticket %s won %s (%s Baht).",
		LossNotification:        "Unfortunately, your ticket %s did not win this time. Better luck next time!",
		DrawWinDetail:           "• Number %s (%s)\n  - %s x%d ticket(s)\n  - Prize: %s Baht",
		DrawLoseDetail:          "• Number %s (%s) x%d ticket(s)",
		DrawWinMessage:          "🏆 Congratulations! You won the Thai Government Lottery\nDraw Date: %s\n\nWinning tickets:\n%s\n\nTotal Prize Amount: %s Baht 🎉",
		DrawJackpotInfo:         "\n\n--------------------\nℹ️ Special 3-Digit Prize (N3 Jackpot)\nSpecial prize number is %s (Prize amount: %s Baht)\nIf your 12-digit number on Pao Tang app matches this number, you won the Special Prize!\n--------------------",
		DrawFootnote:            "\n\n*Please verify official results again for correctness.*",
		DrawLoseMessage:         "Draw Date: %s\n\nUnfortunately, you didn't win any prize this time 😢\nChecked tickets:\n%s\n\nBetter luck next time! ✌️",
		GenericError:            "An error occurred. Please try again 🙏",
		WelcomeGreetingGeneric:  "👋 Hello!",
		WelcomeGreetingPersonal: "👋 Hello %s!",
		DbMaintenance:           "Sorry for the inconvenience. Our system database is temporarily undergoing maintenance. Please try again later.",
		OcrConfirmHeader:        "We detected the following tickets in your photo. Are these correct? 🎟️\n%s",
		OcrWarningHeader:        "\n\n⚠️ Warnings:\n%s",
		OcrConfirmSave:          "Confirm & Save",
		OcrEdit:                 "Edit Numbers",
		OcrCancel:               "Cancel",
		OcrSaved:                "Successfully saved tickets! ✅",
		OcrCancelled:            "Registration cancelled. ❌",
		OcrEditInstruction:      "Please select a ticket number to edit:",
		OcrPromptCorrection:     "Please send the correct number and quantity for %s (e.g. 123456x2 or 123456).",
		OcrInvalidCorrection:    "Invalid input. Please send a valid number (3 or 6 digits) and quantity (e.g. 123456x2 or just 123456).",
		OcrConflictWarning:      "You already have a pending photo registration session. Please complete or cancel it before uploading a new photo.",
		OcrShowPending:          "Show Pending",
		OcrConfirmPending:       "Confirm Pending",
		OcrCancelPending:        "Cancel Pending",
		OcrBack:                 "Back",
	},
	TH: {
		WelcomeFirstTime:        "🎟️ ยินดีต้อนรับสู่ Lotto Journal!\nพิมพ์เลขสลากที่คุณซื้อไว้เพื่อรอตรวจผลอัตโนมัติได้เลย\nตัวอย่าง: 123456 หรือ 456\nส่งหลายเลขได้ เช่น 123456 789012\nระบุจำนวนตั๋วด้วย x เช่น 123456x2\n\n📝 หากต้องการดูสลากที่บันทึกไว้ พิมพ์ 'โพย'\n\n(พิมพ์ `english` เพื่อเปลี่ยนเป็นภาษาอังกฤษ)",
		WelcomeReturning:        "ยินดีต้อนรับกลับสู่ Lotto Journal! ส่งเลขสลากของคุณเพื่อเริ่มต้นบันทึกได้เลย หรือพิมพ์ 'โพย' เพื่อดูสลากที่บันทึกไว้",
		SubmitConfirm:           "บันทึกสลากเรียบร้อย ✅:\n%s",
		SubmitInvalid:           "ไม่พบเลขสลากที่ถูกต้อง ❌\nกรุณาส่งเลข 3 หรือ 6 หลักเท่านั้น\nเลขที่ไม่ถูกต้อง: %s",
		SubmitMixed:             "บันทึกสลากเรียบร้อย ✅:\n%s\n\nเลขที่ไม่ถูกต้อง (ข้ามไป): %s",
		ListHeader:              "สลากที่คุณบันทึกไว้ในงวดนี้ (%s) 📝:",
		ListEmpty:               "คุณยังไม่ได้บันทึกสลากในงวดนี้",
		LanguageSwitched:        "เปลี่ยนภาษาเป็นภาษาไทยเรียบร้อย 🇹🇭",
		AddHelp:                 "🎟️ คุณสามารถบันทึกสลากได้ง่ายๆ โดยพิมพ์ส่งเข้ามาในแชท:\n\n- เลขตัวเดียว: `123456` หรือ `456`\n- ส่งหลายเลขพร้อมกัน: `123456, 654321` หรือ `123 456`\n- ระบุจำนวน: ต่อท้ายด้วย `xN` เช่น `123456x2` (บันทึก 2 ใบ)",
		NotifyHelp:              "🔔 ระบบจะส่งผลรางวัลให้คุณทราบโดยอัตโนมัติทันทีหลังจากประกาศผลรางวัลเสร็จสิ้น (ประมาณ 16:00 น. ของวันที่ 1 และ 16 ของทุกเดือน)",
		WinNotification:         "ยินดีด้วย! สลากเลข %s ของคุณ ถูกรางวัล %s (%s บาท) 🎉",
		LossNotification:        "ขออภัย สลากเลข %s ของคุณ ไม่ถูกรางวัลในงวดนี้ พยายามใหม่อีกครั้งนะ! ✌️",
		DrawWinDetail:           "• เลข %s (%s)\n  - %s x%d ใบ\n  - เงินรางวัล %s บาท",
		DrawLoseDetail:          "• เลข %s (%s) x%d ใบ",
		DrawWinMessage:          "🏆 ยินดีด้วย! คุณถูกรางวัลสลากกินแบ่งรัฐบาล\nงวดประจำวันที่ %s\n\nสลากที่ถูกรางวัล:\n%s\n\nยอดเงินรางวัลรวมทั้งหมด: %s บาท 🎉",
		DrawJackpotInfo:         "\n\n--------------------\nℹ️ ลุ้นรางวัลพิเศษสามตัวท้าย (N3 Jackpot)\nเลขรางวัลพิเศษงวดนี้คือ %s (เงินรางวัล %s บาท)\nหากหมายเลข 12 หลักบนแอปเป๋าตังของคุณตรงกับเลขนี้ คุณคือผู้ถูกรางวัลพิเศษ!\n--------------------",
		DrawFootnote:            "\n\n*กรุณาตรวจสอบผลรางวัลอย่างเป็นทางการอีกครั้งเพื่อความถูกต้อง*",
		DrawLoseMessage:         "งวดประจำวันที่: %s\n\nเสียใจด้วยครับ คุณไม่ถูกรางวัลในงวดนี้ 😢\nสลากที่ตรวจสอบ:\n%s\n\nพยายามใหม่งวดหน้าครับ! ✌️",
		GenericError:            "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง 🙏",
		WelcomeGreetingGeneric:  "👋 สวัสดี!",
		WelcomeGreetingPersonal: "👋 สวัสดีคุณ %s!",
		DbMaintenance:           "ขออภัยในความไม่สะดวก ขณะนี้ระบบฐานข้อมูลอยู่ระหว่างการปรับปรุงชั่วคราว กรุณาลองใหม่อีกครั้งในภายหลัง",
		OcrConfirmHeader:        "เราพบเลขสลากดังต่อไปนี้จากรูปภาพของคุณ ถูกต้องหรือไม่? 🎟️\n%s",
		OcrWarningHeader:        "\n\n⚠️ ข้อควรระวัง:\n%s",
		OcrConfirmSave:          "ยืนยันและบันทึก",
		OcrEdit:                 "แก้ไขข้อมูล",
		OcrCancel:               "ยกเลิก",
		OcrSaved:                "บันทึกสลากเรียบร้อยแล้ว! ✅",
		OcrCancelled:            "ยกเลิกการบันทึกสลากแล้ว ❌",
		OcrEditInstruction:      "กรุณาเลือกเลขที่ต้องการแก้ไข:",
		OcrPromptCorrection:     "กรุณาส่งเลขและจำนวนที่ถูกต้องสำหรับ %s (เช่น 123456x2 หรือ 123456)",
		OcrInvalidCorrection:    "รูปแบบไม่ถูกต้อง กรุณาส่งเลขที่ถูกต้อง (3 หรือ 6 หลัก) พร้อมจำนวน (เช่น 123456x2 หรือ 123456)",
		OcrConflictWarning:      "คุณมีรายการบันทึกสลากจากรูปภาพที่ค้างอยู่ กรุณายืนยันหรือยกเลิกรายการเดิมก่อนส่งรูปภาพใหม่",
		OcrShowPending:          "ดูสลากเดิม",
		OcrConfirmPending:       "ยืนยันสลากเดิม",
		OcrCancelPending:        "ยกเลิกสลากเดิม",
		OcrBack:                 "ย้อนกลับ",
	},
}

const DbMaintenanceMessage = "ขออภัยในความไม่สะดวก ขณะนี้ระบบฐานข้อมูลอยู่ระหว่างการปรับปรุงชั่วคราว กรุณาลองใหม่อีกครั้งในภายหลัง\n\nSorry for the inconvenience. The system is currently undergoing database maintenance. Please try again later."

func GetDictionary(lang string) *Dictionary {
	l := Language(strings.ToLower(lang))
	if dict, ok := Dictionaries[l]; ok {
		return dict
	}
	return Dictionaries[EN] // Default to English
}

func GetQuickReplies(lang string) *messaging_api.QuickReply {
	if strings.ToLower(lang) == "th" {
		return &messaging_api.QuickReply{
			Items: []messaging_api.QuickReplyItem{
				{
					Type: "action",
					Action: &messaging_api.MessageAction{
						Label: "📝 โพย",
						Text:  "โพย",
					},
				},
				{
					Type: "action",
					Action: &messaging_api.MessageAction{
						Label: "🎟️ เพิ่ม",
						Text:  "เพิ่ม",
					},
				},
				{
					Type: "action",
					Action: &messaging_api.MessageAction{
						Label: "🔔 แจ้งเตือน",
						Text:  "แจ้งเตือน",
					},
				},
				{
					Type: "action",
					Action: &messaging_api.MessageAction{
						Label: "🇺🇸 Change Language",
						Text:  "english",
					},
				},
			},
		}
	}

	return &messaging_api.QuickReply{
		Items: []messaging_api.QuickReplyItem{
			{
				Type: "action",
				Action: &messaging_api.MessageAction{
					Label: "📝 List",
					Text:  "list",
				},
			},
			{
				Type: "action",
				Action: &messaging_api.MessageAction{
					Label: "🎟️ Add",
					Text:  "add",
				},
			},
			{
				Type: "action",
				Action: &messaging_api.MessageAction{
					Label: "🔔 Notify",
					Text:  "notify",
				},
			},
			{
				Type: "action",
				Action: &messaging_api.MessageAction{
					Label: "🇹🇭 เปลี่ยนภาษา",
					Text:  "ไทย",
				},
			},
		},
	}
}

func GetPrizeName(category string, lang string) string {
	if strings.ToLower(lang) == "th" {
		switch category {
		case "l6_first":
			return "รางวัลที่ 1"
		case "l6_second":
			return "รางวัลที่ 2"
		case "l6_third":
			return "รางวัลที่ 3"
		case "l6_fourth":
			return "รางวัลที่ 4"
		case "l6_fifth":
			return "รางวัลที่ 5"
		case "l6_last2":
			return "เลขท้าย 2 ตัว"
		case "l6_last3f":
			return "เลขหน้า 3 ตัว"
		case "l6_last3b":
			return "เลขท้าย 3 ตัว"
		case "l6_near_first":
			return "รางวัลข้างเคียงรางวัลที่ 1"
		case "n3_straight_three":
			return "3 ตัวตรง"
		case "n3_shuffle":
			return "3 ตัวสลับ (โต๊ด)"
		case "n3_straight_two":
			return "2 ตัวตรง"
		case "n3_special":
			return "รางวัลพิเศษ (แจ็กพอต)"
		default:
			return category
		}
	}

	switch category {
	case "l6_first":
		return "1st Prize"
	case "l6_second":
		return "2nd Prize"
	case "l6_third":
		return "3rd Prize"
	case "l6_fourth":
		return "4th Prize"
	case "l6_fifth":
		return "5th Prize"
	case "l6_last2":
		return "Last 2-Digit Prize"
	case "l6_last3f":
		return "First 3-Digit Prize"
	case "l6_last3b":
		return "Last 3-Digit Prize"
	case "l6_near_first":
		return "First Prize Neighbors"
	case "n3_straight_three":
		return "3 Straight"
	case "n3_shuffle":
		return "3 Todd (Shuffle)"
	case "n3_straight_two":
		return "2 Straight"
	case "n3_special":
		return "Special Prize (Jackpot)"
	default:
		return category
	}
}

func GetOcrWarningMessage(warning string, lang string) string {
	if strings.ToLower(lang) == "th" {
		switch warning {
		case "image_blurry":
			return "รูปภาพดูเบลอเล็กน้อย"
		case "number_partially_hidden":
			return "เลขบางส่วนอาจถูกบังหรือหลุดขอบรูปภาพ"
		case "multiple_tickets_detected":
			return "ตรวจพบสลากหลายใบ (แนะนำไม่เกิน 5 ใบต่อรูป)"
		case "low_confidence":
			return "ความแม่นยำต่ำ: เลขบางหลักอาจไม่ถูกต้อง"
		case "not_lottery_ticket":
			return "รูปภาพนี้ดูไม่เหมือนสลากกินแบ่งรัฐบาล"
		default:
			return warning
		}
	}
	switch warning {
	case "image_blurry":
		return "The photo seems blurry."
	case "number_partially_hidden":
		return "Some numbers might be partially covered or cropped."
	case "multiple_tickets_detected":
		return "Multiple tickets detected (recommend <= 5 per photo)."
	case "low_confidence":
		return "Low confidence: Some digits might be incorrect."
	case "not_lottery_ticket":
		return "This does not look like a Thai Government Lottery ticket."
	default:
		return warning
	}
}
