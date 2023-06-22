package models

type Campsite struct {
	SiteID    string
	SiteType  string
	MaxPeople int
	MinPeople int
	UseType   string
	Loop      string
}

// Equal checks if the current campsite is equal to another one.
func (c *Campsite) Equal(other Campsite) bool {
	return c.SiteID == other.SiteID &&
		c.SiteType == other.SiteType &&
		c.MaxPeople == other.MaxPeople &&
		c.MinPeople == other.MinPeople &&
		c.UseType == other.UseType &&
		c.Loop == other.Loop
}

// NewCampsite is a constructor for the Campsite struct
func NewCampsite(siteID, siteType, useType, loop string, maxPeople, minPeople int) Campsite {
	return Campsite{
		SiteID:    siteID,
		SiteType:  siteType,
		MaxPeople: maxPeople,
		MinPeople: minPeople,
		UseType:   useType,
		Loop:      loop,
	}
}
