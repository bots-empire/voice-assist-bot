package services

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
	Rows []Rows
}

type Rows interface {
	build(lang string) []tgbotapi.InlineKeyboardButton
}

func NewInlineMarkUp(rows ...Rows) InlineMarkUp {
	return InlineMarkUp{
		Rows: rows,
	}
}

type InlineRow struct {
	Buttons []ButtonData
}

func NewInlineDataRow(buttons ...ButtonData) InlineRow {
	return InlineRow{
		Buttons: buttons,
	}
}

type ButtonData struct {
	Text string
	Data string
}

func NewDataButton(text, data string) ButtonData {
	return ButtonData{
		Text: text,
		Data: data,
	}
}

func (m InlineMarkUp) Build(lang string) tgbotapi.InlineKeyboardMarkup {
	var replyMarkUp tgbotapi.InlineKeyboardMarkup

	for _, elem := range m.Rows {
		replyMarkUp.InlineKeyboard = append(replyMarkUp.InlineKeyboard,
			elem.build(lang))
	}
	return replyMarkUp
}

func (r InlineRow) build(lang string) []tgbotapi.InlineKeyboardButton {
	var replyRow []tgbotapi.InlineKeyboardButton

	for _, elem := range r.Buttons {
		button := tgbotapi.NewInlineKeyboardButtonData(assets.LangText(lang, elem.Text), elem.Data)
		replyRow = append(replyRow, button)
	}
	return replyRow
}

type InlineURLRow struct {
	URLButtons []URLButton
}

func NewInlineURLRow(buttons ...URLButton) InlineURLRow {
	return InlineURLRow{
		URLButtons: buttons,
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

func (r InlineURLRow) build(lang string) []tgbotapi.InlineKeyboardButton {
	var replyRow []tgbotapi.InlineKeyboardButton

	for _, elem := range r.URLButtons {
		button := tgbotapi.NewInlineKeyboardButtonURL(assets.LangText(lang, elem.Text), elem.Url)
		replyRow = append(replyRow, button)
	}
	return replyRow
}
