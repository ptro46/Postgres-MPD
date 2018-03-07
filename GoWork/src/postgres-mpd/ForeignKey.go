package main

import (
	"errors"
	"regexp"
	"strings"
)

type ForeignKey struct {
	Name                 string
	ForeignKeyConstraint string
	ColumnName           string
	RefTable             string
	RefColumn            string
	DeleteCascade        bool
}

func NewForeignKey(name_ string, constraint_ string) *ForeignKey {
	return &ForeignKey{Name: name_, ForeignKeyConstraint: constraint_}
}

// FOREIGN KEY (image_id) REFERENCES media(id) ON DELETE CASCADE
func (foreignKey *ForeignKey) parseWithDeleteCascade() error {
	s := strings.Replace(foreignKey.ForeignKeyConstraint, "(", "'", -1)
	s = strings.Replace(s, ")", "'", -1)
	re := regexp.MustCompile("FOREIGN KEY '([a-zA-Z0-9_-]+)' REFERENCES ([a-zA-Z0-9_-]+)'([a-zA-Z0-9_-]+)' ON DELETE CASCADE")
	matchArray := re.FindSubmatch([]byte(s))
	if len(matchArray) == 4 {
		foreignKey.ColumnName = string(matchArray[1])
		foreignKey.RefTable = string(matchArray[2])
		foreignKey.RefColumn = string(matchArray[3])
	} else {
		return errors.New("Can not parse [" + foreignKey.ForeignKeyConstraint + "]")
	}
	return nil
}

// FOREIGN KEY (image_id) REFERENCES media(id) ON DELETE CASCADE
func (foreignKey *ForeignKey) parseWithoutDeleteCascade() error {
	s := strings.Replace(foreignKey.ForeignKeyConstraint, "(", "'", -1)
	s = strings.Replace(s, ")", "'", -1)
	re := regexp.MustCompile("FOREIGN KEY '([a-zA-Z0-9_-]+)' REFERENCES ([a-zA-Z0-9_-]+)'([a-zA-Z0-9_-]+)'")
	matchArray := re.FindSubmatch([]byte(s))
	if len(matchArray) == 4 {
		foreignKey.ColumnName = string(matchArray[1])
		foreignKey.RefTable = string(matchArray[2])
		foreignKey.RefColumn = string(matchArray[3])
	} else {
		return errors.New("Can not parse [" + foreignKey.ForeignKeyConstraint + "]")
	}
	return nil
}

func (foreignKey *ForeignKey) parse() error {
	err := foreignKey.parseWithDeleteCascade()
	if err != nil {
		return foreignKey.parseWithoutDeleteCascade()
	}
	return nil
}
