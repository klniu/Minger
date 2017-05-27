package model

type PlayList struct {
    ID     int
    Audios []Audio
}

type Audio struct {
    ID         int
    PlayListID int
    Order      int // the order in the playlist
    FilePath   string
}
