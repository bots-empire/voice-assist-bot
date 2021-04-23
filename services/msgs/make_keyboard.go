package msgs

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

/*
==================================================
		MarkUp
==================================================
*/

type MarkUp struct {
	Rows []Row
}

func NewMarkUp(rows ...Row) MarkUp {
	return MarkUp{
		Rows: rows,
	}
}

type Row struct {
	Buttons []string
}

func NewRow(buttons ...string) Row {
	return Row{
		Buttons: buttons,
	}
}

func (m MarkUp) Build(lang string) tgbotapi.ReplyKeyboardMarkup {
	var replyMarkUp tgbotapi.ReplyKeyboardMarkup

	for _, elem := range m.Rows {
		replyMarkUp.Keyboard = append(replyMarkUp.Keyboard,
			elem.build(lang))
	}
	replyMarkUp.ResizeKeyboard = true
	return replyMarkUp
}

func (r Row) build(lang string) []tgbotapi.KeyboardButton {
	var replyRow []tgbotapi.KeyboardButton

	for _, text := range r.Buttons {
		button := tgbotapi.NewKeyboardButton(assets.LangText(lang, text))
		replyRow = append(replyRow, button)
	}
	return replyRow
}

/*
==================================================
		InlineMarkUp
==================================================
*/

type InlineMarkUp struct {
	Rows []InlineRow
}

func NewInlineMarkUp(rows ...InlineRow) InlineMarkUp {
	return InlineMarkUp{
		Rows: rows,
	}
}

func (m InlineMarkUp) Build(lang string) tgbotapi.InlineKeyboardMarkup {
	var replyMarkUp tgbotapi.InlineKeyboardMarkup

	for _, row := range m.Rows {
		replyMarkUp.InlineKeyboard = append(replyMarkUp.InlineKeyboard,
			row.buildRow(lang))
	}
	return replyMarkUp
}

type InlineRow struct {
	Buttons []Buttons
}

type Buttons interface {
	build(lang string) tgbotapi.InlineKeyboardButton
}

func NewInlineRow(buttons ...Buttons) InlineRow {
	return InlineRow{
		Buttons: buttons,
	}
}

type DataButton struct {
	Text string
	Data string
}

func NewDataButton(text, data string) DataButton {
	return DataButton{
		Text: text,
		Data: data,
	}
}

type URLButton struct {
	Text string
	Url  string
}

func NewURLButton(text, url string) URLButton {
	return URLButton{
		Text: text,
		Url:  url,
	}
}

func (r InlineRow) buildRow(lang string) []tgbotapi.InlineKeyboardButton {
	var replyRow []tgbotapi.InlineKeyboardButton

	for _, butt := range r.Buttons {
		replyRow = append(replyRow, butt.build(lang))
	}
	return replyRow
}

func (u URLButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, u.Text)
	return tgbotapi.NewInlineKeyboardButtonURL(text, u.Url)
}

func (d DataButton) build(lang string) tgbotapi.InlineKeyboardButton {
	text := assets.LangText(lang, d.Text)
	return tgbotapi.NewInlineKeyboardButtonData(text, d.Data)
}
