import type { Message, ServerMessage } from '$lib/types';

export default function mapToMessage(backEndMessage: ServerMessage): Message {
    let timestampAsDate = new Date(backEndMessage.timestamp);
    let messageDate = timestampAsDate.toDateString();
    let timeStr = timestampAsDate.toTimeString();
    let mappedMessage: Message = {
        timestamp: backEndMessage.timestamp,
        time: `${messageDate} @ ${timeStr.slice(0, 5)}`,
        content: backEndMessage.content,
        reactions: backEndMessage.reactions,
        sender: backEndMessage.sender
    }

    return mappedMessage
}