<script lang="ts">
	import mapToMessage from '$lib/mapper';
	import { onMount } from 'svelte';
	import MessageComponent from './MessageComponent.svelte';
	import type { Message, ServerMessage } from '$lib/types';
	import { Center, Stack, Button, Group, Chip } from '../../node_modules/@svelteuidev/core';

	let loading = true;
	let randomMessage = {} as Message;
	async function getRandomMessage(): Promise<void> {
		priorMessages = [] as Message[];
		subsequentMessages = [] as Message[];
		loading = true;
		randomMessage = {} as Message;
		const responseObject = await fetch('http://localhost:8080/messages/random').then((r) =>
			r.json()
		);

		randomMessage = mapToMessage(responseObject);
		loading = false;
	}

	let contextLoaded = false;
	let priorMessages = [] as Message[];
	let subsequentMessages = [] as Message[];
	async function getContext(timestamp: number) {
		priorMessages = [] as Message[];
		subsequentMessages = [] as Message[];
		contextLoaded = false;
		console.log('Timestamp for context: ', timestamp);

		const response: ServerMessage[] = await fetch(
			`http://localhost:8080/messages/${timestamp}`
		).then((r) => r.json());

		console.log('context response: ', response);
		for (let i in response) {
			let goPrior = response[i].timestamp < timestamp;
			if (goPrior) {
				priorMessages.push(mapToMessage(response[i]));
			} else {
				subsequentMessages.push(mapToMessage(response[i]));
			}
		}

		contextLoaded = true;
	}

	onMount(getRandomMessage);

	let cbSelected = false;
	let ssSelected = false;
	let kkSelected = false;
	let nzSelected = false
	let kwSelected = false;
</script>

<Center>
	<Stack>
		{#if loading}
			<p>Getting random message...</p>
		{/if}

		{#if contextLoaded}
			{#each priorMessages as pMessage}
				<MessageComponent message={pMessage} />
			{/each}
		{/if}
		{#if !loading}
			<MessageComponent message={randomMessage} />
		{/if}
		{#if contextLoaded}
			{#each subsequentMessages as sMessage}
				<MessageComponent message={sMessage} />
			{/each}
		{/if}

		<Group position="center">
			<Chip
				on:change={() => {
					cbSelected = !cbSelected;
				}}
			>
				CB
			</Chip>
			<Chip
				on:change={() => {
					ssSelected = !ssSelected;
				}}
			>
				SS
			</Chip>
			<Chip
				on:change={() => {
					kkSelected = !kkSelected;
				}}
			>
				KK
			</Chip>
			<Chip
				on:change={() => {
					nzSelected = !nzSelected;
				}}
			>
				NZ	
			</Chip>
			<Chip
				on:change={() => {
					kwSelected = !kwSelected;
				}}
			>
				KW
			</Chip>
		</Group>

		<Group position="center">
			<Button on:click={() => getRandomMessage()}>Get Random Message</Button>
			<Button
				disabled={randomMessage.timestamp === undefined}
				on:click={() => getContext(randomMessage.timestamp)}
			>
				Get Context
			</Button>
		</Group>
	</Stack>
</Center>
