package models

type LastVisitedPage struct {
	BaseModel
	Bundle         string `json:"bundle"`
	Pathname       string `json:"pathname"`
	Title          string `json:"title"`
	UserIdentityID uint   `json:"userIdentityId"`
}

type LastVisitedPages struct {
	Pages []LastVisitedPage `json:"pages"`
}

type LastVisitedPageResponse struct {
	Bundle   string `json:"bundle"`
	Pathname string `json:"pathname"`
	Title    string `json:"title"`
}

func CastLastVisitedResponse(pages []LastVisitedPage) []LastVisitedPageResponse {
	var casted []LastVisitedPageResponse
	for _, v := range pages {
		casted = append(casted, LastVisitedPageResponse{
			Bundle:   v.Bundle,
			Pathname: v.Pathname,
			Title:    v.Title,
		})
	}
	return casted
}
