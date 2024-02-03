export type ServerMessage = {
    audio: BackEndAudio[],
    content: string,
    gifs: Gif[],
    id: string,
    photos: Photo[],
    reactions: Reaction[],
    sender: string,
    share: Share,
    timestamp: number,
    type: string,
    videos: Video[]
}

export type Message = {
    audio: BronerAudio[],
    content: string,
    gifs: Gif[],
    photos: Photo[],
    reactions: Reaction[],
    sender: string,
    share: Share,
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
    link: string,
    shareText: string
}

export type Video = {
    uri: string
}