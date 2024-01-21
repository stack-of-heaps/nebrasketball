<script lang="ts">
	import type { Message } from '$lib/types';
	export let message = {} as Message;
</script>

<p>{message.sender}</p>
<p>{message.content}</p>
<p>{message.time}</p>
{#if message.reactions}
	{#each message.reactions as reaction}
		<p>{reaction.actor} {reaction.reaction}</p>
	{/each}
{/if}
