package main

import (
	"github.com/iansmith/hood"
	"nullblog"
)

func (m *M) CreateArticleTable_1375646159_Up(hd *hood.Hood) {
	hd.CreateTable(&nullblog.Article{})
}

func (m *M) CreateArticleTable_1375646159_Down(hd *hood.Hood) {
	//nothing to do here
}
