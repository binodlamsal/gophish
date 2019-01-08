package models

import (
	"errors"
	"time"

	log "github.com/binodlamsal/gophish/logger"
	"github.com/jinzhu/gorm"
)

// Template models hold the attributes for an email template to be sent to targets
type Template struct {
	Id           int64        `json:"id" gorm:"column:id; primary_key:yes"`
	UserId       int64        `json:"-" gorm:"column:user_id"`
	Name         string       `json:"name"`
	Subject      string       `json:"subject"`
	Text         string       `json:"text"`
	HTML         string       `json:"html" gorm:"column:html"`
	RATING       int64        `json:"rating" gorm:"column:rating"`
	TagsId       int64        `json:"tag" gorm:"column:tag"`
	Tags         Tags         `json:"tags"`
	Public       bool         `json:"public" gorm:"column:public"`
	ModifiedDate time.Time    `json:"modified_date"`
	Attachments  []Attachment `json:"attachments"`
}

// Tags models hold the attributes for the categories of templates and landing pages
type Tags struct {
	Id     int64  `json:"id" gorm:"column:id; primary_key:yes"`
	Name   string `json:"name"`
	Weight int64  `json:"weight"`
}

// ErrTemplateNameNotSpecified is thrown when a template name is not specified
var ErrTemplateNameNotSpecified = errors.New("Template name not specified")

// ErrTemplateMissingParameter is thrown when a needed parameter is not provided
var ErrTemplateMissingParameter = errors.New("Need to specify at least plaintext or HTML content")

// Validate checks the given template to make sure values are appropriate and complete
func (t *Template) Validate() error {
	switch {
	case t.Name == "":
		return ErrTemplateNameNotSpecified
	case t.Text == "" && t.HTML == "":
		return ErrTemplateMissingParameter
	}
	if err = ValidateTemplate(t.HTML); err != nil {
		return err
	}
	if err = ValidateTemplate(t.Text); err != nil {
		return err
	}
	return nil
}

//Get tags by tag name
func GetTagById(id int64) (Tags, error) {
	t := Tags{}
	err := db.Where("id=?", id).Find(&t).Error
	if err != nil {
		log.Error(err)
	}
	return t, err
}

// GetTemplates returns the templates owned by the given user.
func GetTemplates(uid int64) ([]Template, error) {
	ts := []Template{}
	err := db.Where("user_id=? OR public=?", uid, 1).Find(&ts).Error
	if err != nil {
		log.Error(err)
		return ts, err
	}
	for i, _ := range ts {
		// Get Attachments
		err = db.Where("template_id=?", ts[i].Id).Find(&ts[i].Attachments).Error
		if err == nil && len(ts[i].Attachments) == 0 {
			ts[i].Attachments = make([]Attachment, 0)
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			return ts, err
		}
	}
	return ts, err
}

// GetTags returns the all the tags from the database
func GetTags(uid int64) ([]Tags, error) {
	tg := []Tags{}
	err := db.Order("id asc").Find(&tg).Error
	return tg, err
}

// PostTemplate creates a new template in the database.
func PostTags(t *Tags) error {
	// Insert into the DB
	if t.Name == "" {
		return errors.New("Tag name is not specified")
	}
	if t.Weight == 0 {
		return errors.New("Weight is not specified")
	}

	err = db.Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// GetTemplate returns the template, if it exists, specified by the given id and user_id.
func GetTemplate(id int64, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&t).Error
	if err != nil {
		log.Error(err)
		return t, err
	}

	// Get Attachments
	err = db.Where("template_id=?", t.Id).Find(&t.Attachments).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return t, err
	}
	if err == nil && len(t.Attachments) == 0 {
		t.Attachments = make([]Attachment, 0)
	}
	return t, err
}

// GetTemplateByName returns the template, if it exists, specified by the given name and user_id.
func GetTemplateByName(n string, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and name=?", uid, n).Or("public = ? and name=?", 1, n).Find(&t).Error
	if err != nil {
		log.Error(err)
		return t, err
	}

	// Get Attachments
	err = db.Where("template_id=?", t.Id).Find(&t.Attachments).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return t, err
	}
	if err == nil && len(t.Attachments) == 0 {
		t.Attachments = make([]Attachment, 0)
	}
	return t, err
}

// PostTemplate creates a new template in the database.
func PostTemplate(t *Template) error {
	// Insert into the DB
	if err := t.Validate(); err != nil {
		return err
	}

	tg, err := GetTagById(t.TagsId)

	if err != nil {
		log.Error(err)
		return err
	}

	t.Tags = tg

	err = db.Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}

	// Save every attachment
	for i := range t.Attachments {
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// PutTemplate edits an existing template in the database.
// Per the PUT Method RFC, it presumes all data for a template is provided.
func PutTemplate(t *Template) error {
	if err := t.Validate(); err != nil {
		return err
	}
	// Delete all attachments, and replace with new ones
	err = db.Where("template_id=?", t.Id).Delete(&Attachment{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return err
	}
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	for i, _ := range t.Attachments {
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}

	// Save final template
	err = db.Where("id=?", t.Id).Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// PutTags edits an existing tag in the database.
// Per the PUT Method RFC, it presumes all data for tag is provided.
func PutTags(t *Tags) error {
	// Save final template
	err = db.Where("id=?", t.Id).Save(t).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// DeleteTemplate deletes an existing template in the database.
// An error is returned if a template with the given user id and template id is not found.
func DeleteTemplate(id int64, uid int64) error {
	// Delete attachments
	err := db.Where("template_id=?", id).Delete(&Attachment{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	// Finally, delete the template itself
	err = db.Where("user_id=?", uid).Delete(Template{Id: id}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// DeleteTags deletes an existing tag in the database.
// An error is returned if a template with the given user id and tag id is not found.
func DeleteTags(id int64) error {
	// Finally, delete the template itself
	err := db.Delete(Tags{Id: id}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
