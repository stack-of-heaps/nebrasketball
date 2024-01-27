<script lang="ts">
	import mapToMessage from '$lib/mapper';
	import { onMount } from 'svelte';
	import MessageComponent from './MessageComponent.svelte';
	import type { Message, ServerMessage } from '$lib/types';
	import { Center, Stack, Button, Group, Chip } from '../../node_modules/@svelteuidev/core';

	let loading = true;
	let messages = [] as Message[];
	let randomMessage = {} as Message;

	async function getRandomMessage(participants: string): Promise<void> {
		loading = true;
		randomMessage = {} as Message;
		messages = [];
		const responseObject = await fetch(
			`http://localhost:8080/messages/random?participants=${participants}`
		).then((r) => r.json());

		let seedMessage = mapToMessage(responseObject);
		randomMessage = seedMessage;
		messages.push(seedMessage);
		messages = messages;
		loading = false;
	}

	let contextLoaded = false;
	async function getContext(timestamp: number) {
		contextLoaded = false;
		console.log('Timestamp for context: ', timestamp);

		const response: ServerMessage[] = await fetch(
			`http://localhost:8080/messages/${timestamp}`
		).then((r) => r.json());

		console.log('context response: ', response);
		for (let i in response) {
			if (messages.find((m) => m.timestamp === response[i].timestamp)) {
				continue;
			} else {
				messages.push(mapToMessage(response[i]));
			}
		}

		messages = messages.sort(
			(firstMessage, secondMessage) => firstMessage.timestamp - secondMessage.timestamp
		);
		contextLoaded = true;
	}

	function getParticipants(): string {
		let participants = '';
		if (cbSelected) participants += 'Charles Baker,';
		if (ssSelected) participants += 'Spencer Smith,';
		if (kkSelected) participants += 'Kyle Karthauser,';
		if (nzSelected) participants += 'Nathan Zielinski,';
		if (kwSelected) participants += 'Kiel Walker,';

		return participants;
	}

	onMount(() => {
		getRandomMessage(getParticipants());
	});

	let cbSelected = false;
	let ssSelected = false;
	let kkSelected = false;
	let nzSelected = false;
	let kwSelected = false;
</script>

<Center>
	<Stack>
		{#if loading}
			<p>Getting random message...</p>
		{/if}

		{#if !loading}
			{#each messages as message}
				<MessageComponent {message} />
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
			<Button on:click={() => getRandomMessage(getParticipants())}>Get Random Message</Button>
			<Button disabled={messages.length != 1} on:click={() => getContext(randomMessage.timestamp)}>
				Get Context
			</Button>
		</Group>
	</Stack>
</Center>
