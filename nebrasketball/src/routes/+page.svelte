<script lang>
	import { onMount } from 'svelte';

	let loading = true;
	let randomMessage = {};
	async function getRandomMessage() {
		loading = true;
		const responseObject = await fetch('http://localhost:8080/messages/random').then((r) =>
			r.json()
		);

		randomMessage = mapToMessage(responseObject);
		loading = false;
	}

	let contextLoaded = false;
	let priorMessages = [];
	let subsequentMessages = [];
	async function getContext(timestamp) {
		contextLoaded = false;
		console.log('Timestamp for context: ', timestamp);

		const response = await fetch(`http://localhost:8080/messages/${timestamp}`).then((r) =>
			r.json()
		);

		console.log('context response: ', response);
		for (let i in response) {
			let goPrior = response[i].Timestamp < timestamp;
			if (goPrior) {
				priorMessages.push(mapToMessage(response[i]));
			} else {
				subsequentMessages.push(mapToMessage(response[i]));
			}
		}

		console.log('Prior: ', priorMessages);
		console.log('Subsequent: ', subsequentMessages);
		contextLoaded = true;
	}

	function mapToMessage(backEndMessage) {
		let timestampAsDate = new Date(backEndMessage.Timestamp);
		let messageDate = timestampAsDate.toDateString();
		let timeStr = timestampAsDate.toTimeString();
		let mappedMessage = {};
		mappedMessage.timestamp = backEndMessage.Timestamp;
		mappedMessage.time = `${messageDate} @ ${timeStr.slice(0, 5)}`;
		mappedMessage.content = backEndMessage.Content;
		mappedMessage.reactions = backEndMessage.Reactions;
		mappedMessage.sender = backEndMessage.Sender;
		mappedMessage.reactions = backEndMessage.Reactions;

		return mappedMessage;
	}

	onMount(getRandomMessage);
</script>

{#if loading}
	<p>Getting random message...</p>
{/if}

{#if contextLoaded}
	{#each priorMessages as pMessage}
		<p>{pMessage.sender}</p>
		<p>{pMessage.content}</p>
		<p>{pMessage.time}</p>
		{#if pMessage.reactions}
			{#each pMessage.reactions as reaction}
				<p>{reaction.Actor} {reaction.Reaction}</p>
			{/each}
		{/if}
	{/each}
{/if}
{#if !loading}
	<p>{randomMessage.sender}</p>
	<p>{randomMessage.content}</p>
	<p>{randomMessage.time}</p>
	{#if randomMessage.reactions}
		{#each randomMessage.reactions as reaction}
			<p>{reaction.Actor} {reaction.Reaction}</p>
		{/each}
	{/if}
{/if}
{#if contextLoaded}
	{#each subsequentMessages as sMessage}
		<p>{sMessage.sender}</p>
		<p>{sMessage.content}</p>
		<p>{sMessage.time}</p>
		{#if sMessage.reactions}
			{#each sMessage.reactions as reaction}
				<p>{reaction.Actor} {reaction.Reaction}</p>
			{/each}
		{/if}
	{/each}
{/if}

<button on:click={() => getRandomMessage()}> Get Random Message </button>
<button
	disabled={randomMessage.timestamp === undefined}
	on:click={() => getContext(randomMessage.timestamp)}
>
	Get Context
</button>
