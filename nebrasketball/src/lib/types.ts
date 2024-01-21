export type ServerMessage = {
    Id: string,
    Sender: string,
    Timestamp: number,
    Content: string,
    Photos: Photo[],
    Reactions: Reaction[],
    Gifs: Gif[],
    Videos: Video[],
    Share: Share,
    Type: string
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