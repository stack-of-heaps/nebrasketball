export type ServerMessage = {
    Id: string,
    sender: string,
    timestamp: number,
    content: string,
    photos: Photo[],
    shares: Share[],
    reactions: Reaction[],
    gifs: Gif[],
    videos: Video[],
    share: Share,
    type: string,
    audio: BackEndAudio[]
}

export type Message = {
    audio: BronerAudio[],
    content: string,
    gifs: Gif[],
    photos: Photo[],
    reactions: Reaction[],
    sender: string,
    time: string,
    timestamp: number,
    videos: Video[]
}

export type BackEndAudio = {
    uri: string
}

export type BronerAudio = {
    uri: string
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