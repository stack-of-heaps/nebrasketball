<script lang="ts">
	import { Group, Chip } from '@svelteuidev/core';

	export let senderSelections = '';

	let cbSelected = false;
	let ssSelected = false;
	let kkSelected = false;
	let nzSelected = false;
	let kwSelected = false;

	function getParticipants(): string {
		let participants = '';
		if (cbSelected) participants += 'cb,';
		if (ssSelected) participants += 'ss,';
		if (kkSelected) participants += 'kk,';
		if (nzSelected) participants += 'nz,';
		if (kwSelected) participants += 'kw,';

		return participants;
	}
</script>

<Group position="center">
	<Chip
		on:change={() => {
			cbSelected = !cbSelected;
			senderSelections = getParticipants();
		}}
	>
		CB
	</Chip>
	<Chip
		on:change={() => {
			ssSelected = !ssSelected;
			senderSelections = getParticipants();
		}}
	>
		SS
	</Chip>
	<Chip
		on:change={() => {
			kkSelected = !kkSelected;
			senderSelections = getParticipants();
		}}
	>
		KK
	</Chip>
	<Chip
		on:change={() => {
			nzSelected = !nzSelected;
			senderSelections = getParticipants();
		}}
	>
		NZ
	</Chip>
	<Chip
		on:change={() => {
			kwSelected = !kwSelected;
			senderSelections = getParticipants();
		}}
	>
		KW
	</Chip>
</Group>
