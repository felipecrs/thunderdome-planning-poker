<script>
    import Modal from '../Modal.svelte'
    import SolidButton from '../SolidButton.svelte'
    import { quill } from '../../quill.js'

    export let toggleCheckin = () => {}
    export let handleCheckin = () => {}
    export let handleCheckinEdit = () => {}
    export let userId
    export let checkinId
    export let today = ''
    export let yesterday = ''
    export let blockers = ''
    export let discuss = ''
    export let goalsMet = true

    function onSubmit(e) {
        e.preventDefault()

        if (checkinId) {
            handleCheckinEdit(checkinId, {
                yesterday,
                today,
                blockers,
                discuss,
                goalsMet,
            })
        } else {
            handleCheckin({
                userId,
                yesterday,
                today,
                blockers,
                discuss,
                goalsMet,
            })
        }
    }
</script>

<style>
    .toggle-checkbox:checked {
        @apply right-0;
        @apply border-green-500;
    }

    .toggle-checkbox:checked + .toggle-label {
        @apply bg-green-500;
    }
</style>

<Modal closeModal="{toggleCheckin}" widthClasses="md:w-2/3">
    <form on:submit="{onSubmit}" name="teamCheckin">
        <div class="mb-4">
            <div
                class="text-gray-700 uppercase font-rajdhani text-2xl tracking-wide mb-2 dark:text-gray-400"
            >
                Did you meet yesterday's goals?
            </div>
            <div
                class="relative inline-block w-16 mr-2 align-middle select-none transition duration-200 ease-in"
            >
                <input
                    type="checkbox"
                    name="goalsMet"
                    id="goalsMet"
                    bind:checked="{goalsMet}"
                    class="toggle-checkbox absolute block w-8 h-8 rounded-full bg-white border-4 border-gray-300 appearance-none cursor-pointer transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 shadow"
                />
                <label
                    for="goalsMet"
                    class="toggle-label block overflow-hidden h-8 rounded-full bg-gray-300 cursor-pointer transition-colors duration-200 ease-in-out"
                >
                </label>
            </div>
        </div>

        <div class="mb-4">
            <div
                class="text-blue-500 uppercase font-rajdhani tracking-wide text-2xl mb-2"
            >
                Yesterday
            </div>
            <div class="bg-white">
                <div
                    class="w-full"
                    use:quill="{{
                        placeholder: `Yesterday I...`,
                        content: yesterday,
                    }}"
                    on:text-change="{e => (yesterday = e.detail.html)}"
                    id="yesterday"
                ></div>
            </div>
        </div>

        <div class="mb-4">
            <div
                class="text-green-500 uppercase font-rajdhani tracking-wide text-2xl mb-2"
            >
                Today
            </div>
            <div class="bg-white">
                <div
                    class="w-full"
                    use:quill="{{
                        placeholder: `Today I will...`,
                        content: today,
                    }}"
                    on:text-change="{e => (today = e.detail.html)}"
                    id="today"
                ></div>
            </div>
        </div>

        <div class="mb-4">
            <div
                class="text-red-500 uppercase font-rajdhani tracking-wide text-2xl mb-2"
            >
                Blockers
            </div>
            <div class="bg-white">
                <div
                    class="w-full"
                    use:quill="{{
                        placeholder: `I'm blocked by...`,
                        content: blockers,
                    }}"
                    on:text-change="{e => (blockers = e.detail.html)}"
                    id="blockers"
                ></div>
            </div>
        </div>

        <div class="mb-4">
            <div
                class="text-purple-500 uppercase font-rajdhani tracking-wide text-2xl mb-2"
            >
                Discuss
            </div>
            <div class="bg-white">
                <div
                    class="w-full"
                    use:quill="{{
                        placeholder: 'I would like to discuss...',
                        content: discuss,
                    }}"
                    on:text-change="{e => (discuss = e.detail.html)}"
                    id="discuss"
                ></div>
            </div>
        </div>

        <div>
            <div class="text-right">
                <SolidButton type="submit">Save</SolidButton>
            </div>
        </div>
    </form>
</Modal>
