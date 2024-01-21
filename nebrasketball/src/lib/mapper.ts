import { Message, ServerMessage } 

export default function mapToMessage(backEndMessage: ServerMessage): Message {
    let timestampAsDate = new Date(backEndMessage.Timestamp);
    let messageDate = timestampAsDate.toDateString();
    let timeStr = timestampAsDate.toTimeString();
    let mappedMessage = {} as Message;
    mappedMessage.timestamp = backEndMessage.Timestamp;
    mappedMessage.time = `${messageDate} @ ${timeStr.slice(0, 5)}`;
    mappedMessage.content = backEndMessage.Content;
    mappedMessage.reactions = backEndMessage.Reactions;
    mappedMessage.sender = backEndMessage.Sender;
    mappedMessage.reactions = backEndMessage.Reactions;

    return mappedMessage;
}