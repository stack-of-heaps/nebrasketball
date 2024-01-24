import type { Message, ServerMessage, Reaction } from '$lib/types';

export default function mapToMessage(backEndMessage: ServerMessage): Message {
    let timestampAsDate = new Date(backEndMessage.timestamp);
    let messageDate = timestampAsDate.toDateString();
    let timeStr = timestampAsDate.toTimeString();
    let mappedMessage: Message = {
        audio: backEndMessage.audio,
        content: backEndMessage.content,
        gifs: backEndMessage.gifs,
        photos: backEndMessage.photos,
        reactions: backEndMessage.reactions.map((r) => {
            if (!r) {
                return {} as Reaction
            }
            return { actor: getInitials(r.actor), reaction: r.reaction } as Reaction
        }),
        sender: getInitials(backEndMessage.sender),
        time: `${messageDate} @ ${timeStr.slice(0, 5)}`,
        timestamp: backEndMessage.timestamp,
        videos: backEndMessage.videos
    }

    return mappedMessage
}

function getInitials(name: string): string {
    let splitName = name.split(' ')
    return `${splitName[0][0]}${splitName[1][0]}`
}