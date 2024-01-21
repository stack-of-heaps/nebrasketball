export type ServerMessage = {
    Id: string,
    sender: string,
    timestamp: number,
    content: string,
    photos: Photo[],
    reactions: Reaction[],
    gifs: Gif[],
    videos: Video[],
    share: Share,
    type: string
}

export type Message = {
    content: string,
    sender: string,
    timestamp: number,
    time: string,
    reactions: Reaction[]
}

export type Reaction = {
    actor: string,
    reaction: string
}

export type Gif = {
    uri: string
}

export type Photo = {
    uri: string
}

export type Share = {
    Link: string,
    ShareText: string
}

export type Video = {
    uri: string
}