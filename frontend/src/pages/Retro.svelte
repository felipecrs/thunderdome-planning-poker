<script>
    import Sockette from 'sockette'
    import { onDestroy, onMount } from 'svelte'

    import PageLayout from '../components/PageLayout.svelte'
    import UserCard from '../components/retro/UserCard.svelte'
    import InviteUser from '../components/retro/InviteUser.svelte'
    import UsersIcon from '../components/icons/Users.svelte'
    import HollowButton from '../components/HollowButton.svelte'
    import DownCarrotIcon from '../components/icons/ChevronDown.svelte'
    import ChevronRight from '../components/icons/ChevronRight.svelte'
    import DeleteConfirmation from '../components/DeleteConfirmation.svelte'
    import SolidButton from '../components/SolidButton.svelte'
    import CheckCircle from '../components/icons/CheckCircle.svelte'
    import CheckboxIcon from '../components/icons/CheckboxIcon.svelte'
    import TrashIcon from '../components/icons/TrashIcon.svelte'
    import RetroItemForm from '../components/retro/ItemForm.svelte'
    import GroupPhase from '../components/retro/GroupPhase.svelte'
    import VotePhase from '../components/retro/VotePhase.svelte'
    import Export from '../components/retro/Export.svelte'
    import ExternalLinkIcon from '../components/icons/ExternalLinkIcon.svelte'
    import { appRoutes, PathPrefix } from '../config'
    import { warrior as user } from '../stores.js'
    import { _ } from '../i18n'

    export let retroId
    export let notifications
    export let router
    export let eventTag

    const { AllowRegistration } = appConfig
    const loginOrRegister = AllowRegistration
        ? appRoutes.register
        : appRoutes.login

    const hostname = window.location.origin
    const socketExtension = window.location.protocol === 'https:' ? 'wss' : 'ws'

    let socketError = false
    let socketReconnecting = false
    let retro = {
        ownerId: '',
        phase: 'brainstorm',
        users: [],
        items: [],
        groups: [],
        actionItems: [],
        votes: [],
    }
    let showUsers = false
    let showDeleteRetro = false
    let actionItem = ''
    let showExport = false
    let groupedItems = []
    let JoinPassRequired = false
    let joinPasscode = ''
    let voteLimitReached = false

    function organizeItemsByGroup() {
        const groupMap = retro.groups.reduce((prev, g) => {
            prev[g.id] = {
                id: g.id,
                name: g.name,
                items: [],
                votes: [],
                userVoted: false,
            }
            return prev
        }, {})
        let userVoteCount = 0

        retro.items.map(item => {
            groupMap[item.groupId].items.push(item)
        })

        retro.votes.map(vote => {
            groupMap[vote.groupId].votes.push(vote.userId)
            if (vote.userId === $user.id) {
                ++userVoteCount
                groupMap[vote.groupId].userVoted = true
            }
        })

        voteLimitReached = userVoteCount === 3

        return Object.values(groupMap)
    }

    const onSocketMessage = function (evt) {
        const parsedEvent = JSON.parse(evt.data)

        switch (parsedEvent.type) {
            case 'join_code_required':
                JoinPassRequired = true
                break
            case 'join_code_incorrect':
                notifications.danger($_('incorrectPassCode'))
                break
            case 'init':
                JoinPassRequired = false
                retro = JSON.parse(parsedEvent.value)
                if (retro.phase != 'brainstorm') {
                    groupedItems = organizeItemsByGroup()
                }
                eventTag('join', 'retro', '')
                break
            case 'user_joined': {
                retro.users = JSON.parse(parsedEvent.value) || []
                const joinedUser = retro.users.find(
                    u => u.id === parsedEvent.userId,
                )
                notifications.success(`${joinedUser.name} joined.`)
                break
            }
            case 'user_left': {
                const leftUser = retro.users.find(
                    w => w.id === parsedEvent.userId,
                )
                retro.users = JSON.parse(parsedEvent.value)

                notifications.danger(`${leftUser.name} retreated.`)
                break
            }
            case 'retro_updated':
                retro = JSON.parse(parsedEvent.value)
                groupedItems = organizeItemsByGroup()
                break
            case 'items_updated': {
                const parsedValue = JSON.parse(parsedEvent.value)
                retro.items = parsedValue
                if (retro.phase !== 'brainstorm') {
                    groupedItems = organizeItemsByGroup()
                }
                break
            }
            case 'groups_updated': {
                const parsedValue = JSON.parse(parsedEvent.value)
                retro.groups = parsedValue
                groupedItems = organizeItemsByGroup()
                break
            }
            case 'votes_updated': {
                const parsedValue = JSON.parse(parsedEvent.value)
                retro.votes = parsedValue
                if (retro.phase === 'vote') {
                    groupedItems = organizeItemsByGroup()
                }
                break
            }
            case 'action_updated':
                retro.actionItems = JSON.parse(parsedEvent.value)
                break
            case 'conceded':
                // retro over, goodbye.
                notifications.warning('Retro deleted')
                router.route(appRoutes.retros)
                break
            default:
                break
        }
    }

    const ws = new Sockette(
        `${socketExtension}://${window.location.host}${PathPrefix}/api/retro/${retroId}`,
        {
            timeout: 2e3,
            maxAttempts: 15,
            onmessage: onSocketMessage,
            onerror: () => {
                socketError = true
                eventTag('socket_error', 'retro', '')
            },
            onclose: e => {
                if (e.code === 4004) {
                    eventTag('not_found', 'retro', '', () => {
                        router.route(appRoutes.retros)
                    })
                } else if (e.code === 4001) {
                    eventTag('socket_unauthorized', 'retro', '', () => {
                        user.delete()
                        router.route(`${appRoutes.register}/retro/${retroId}`)
                    })
                } else if (e.code === 4003) {
                    eventTag('socket_duplicate', 'retro', '', () => {
                        notifications.danger(
                            `Duplicate retro session exists for your ID`,
                        )
                        router.route(`${appRoutes.retros}`)
                    })
                } else if (e.code === 4002) {
                    eventTag('retro_user_abandoned', 'retro', '', () => {
                        router.route(appRoutes.retros)
                    })
                } else {
                    socketReconnecting = true
                    eventTag('socket_close', 'retro', '')
                }
            },
            onopen: () => {
                socketError = false
                socketReconnecting = false
                eventTag('socket_open', 'retro', '')
            },
            onmaximum: () => {
                socketReconnecting = false
                eventTag(
                    'socket_error',
                    'retro',
                    'Socket Reconnect Max Reached',
                )
            },
        },
    )

    onDestroy(() => {
        eventTag('leave', 'retro', '', () => {
            ws.close()
        })
    })

    $: isOwner = retro.ownerId === $user.id

    const sendSocketEvent = (type, value) => {
        ws.send(
            JSON.stringify({
                type,
                value,
            }),
        )
    }

    function concedeRetro() {
        eventTag('concede', 'retro', '', () => {
            sendSocketEvent('concede_retro', '')
        })
    }

    function abandonRetro() {
        eventTag('abandon', 'retro', '', () => {
            sendSocketEvent('abandon_retro', '')
        })
    }

    function toggleUsersPanel() {
        showUsers = !showUsers
        eventTag('show_users', 'retro', `show: ${showUsers}`)
    }

    const toggleDeleteRetro = () => {
        showDeleteRetro = !showDeleteRetro
    }

    const toggleExport = () => {
        showExport = !showExport
    }

    const handleItemAdd = (type, content) => {
        sendSocketEvent(
            `create_item`,
            JSON.stringify({
                type,
                content,
                phase: retro.phase,
            }),
        )
    }

    const handleItemGroupChange = (itemId, groupId) => {
        sendSocketEvent(
            `group_item`,
            JSON.stringify({
                itemId,
                groupId,
            }),
        )
    }

    const handleGroupNameChange = (groupId, name) => {
        sendSocketEvent(
            `group_name_change`,
            JSON.stringify({
                groupId,
                name,
            }),
        )
    }

    const handleItemDelete = (type, id) => () => {
        sendSocketEvent(
            `delete_item`,
            JSON.stringify({
                id,
                type,
                phase: retro.phase,
            }),
        )
    }

    const handleActionItem = evt => {
        evt.preventDefault()

        sendSocketEvent(
            'create_action',
            JSON.stringify({
                content: actionItem,
            }),
        )
        actionItem = ''
    }

    const handleActionUpdate = (id, completed, content) => evt => {
        sendSocketEvent(
            'update_action',
            JSON.stringify({
                id,
                completed: !completed,
                content,
            }),
        )
    }

    const handleActionDelete = id => () => {
        sendSocketEvent(
            'delete_action',
            JSON.stringify({
                id,
            }),
        )
    }

    const advancePhase = () => {
        const nextPhase = {
            intro: 'brainstorm',
            brainstorm: 'group',
            group: 'vote',
            vote: 'action',
            action: 'completed',
        }

        sendSocketEvent(
            'advance_phase',
            JSON.stringify({
                phase: nextPhase[retro.phase],
            }),
        )
    }

    const handleVote = groupId => {
        sendSocketEvent(
            `group_vote`,
            JSON.stringify({
                groupId,
            }),
        )
    }

    const handleVoteSubtract = groupId => {
        sendSocketEvent(
            `group_vote_subtract`,
            JSON.stringify({
                groupId,
            }),
        )
    }

    function authRetro(e) {
        e.preventDefault()

        sendSocketEvent('auth_retro', joinPasscode)
        eventTag('auth_retro', 'retro', '')
    }

    onMount(() => {
        if (!$user.id) {
            router.route(`${loginOrRegister}/retro/${retroId}`)
        }
    })

    $: workedItems =
        retro.items &&
        retro.items.reduce((prev, item) => {
            if (item.type === 'worked') {
                prev.push(item)
            }
            return prev
        }, [])
    $: improveItems =
        retro.items &&
        retro.items.reduce((prev, item) => {
            if (item.type === 'improve') {
                prev.push(item)
            }
            return prev
        }, [])
    $: questionItems =
        retro.items &&
        retro.items.reduce((prev, item) => {
            if (item.type === 'question') {
                prev.push(item)
            }
            return prev
        }, [])
</script>

<style>
    :global(input:checked ~ div) {
        @apply border-green-500;
    }

    :global(input:checked ~ div svg) {
        @apply block;
    }
</style>

<svelte:head>
    <title>Retro {retro.name} | Thunderdome</title>
</svelte:head>

{#if retro.name && !socketReconnecting && !socketError}
    <div
        class="px-6 py-2 bg-gray-100 dark:bg-gray-800 border-b border-t border-gray-400 dark:border-gray-700 flex
        flex-wrap"
    >
        <div class="w-1/4">
            <h1 class="text-3xl font-bold leading-tight dark:text-gray-200">
                {retro.name}
            </h1>
        </div>
        <div class="w-3/4 text-right">
            <div>
                {#if retro.phase === 'completed'}
                    <SolidButton color="green" onClick="{toggleExport}">
                        {#if showExport}
                            Back
                        {:else}
                            Export
                        {/if}
                    </SolidButton>
                {/if}
                {#if isOwner}
                    {#if retro.phase !== 'completed'}
                        <SolidButton color="blue" onClick="{advancePhase}">
                            Next Phase
                        </SolidButton>
                    {/if}

                    <HollowButton
                        color="red"
                        onClick="{toggleDeleteRetro}"
                        class="mr-2"
                    >
                        Delete Retro
                    </HollowButton>
                {:else}
                    <HollowButton color="red" onClick="{abandonRetro}">
                        Leave Retro
                    </HollowButton>
                {/if}
                <div class="inline-block relative">
                    <HollowButton
                        color="teal"
                        class="transition ease-in-out duration-150"
                        onClick="{toggleUsersPanel}"
                    >
                        <UsersIcon />&nbsp; Users
                        <DownCarrotIcon />
                    </HollowButton>
                    {#if showUsers}
                        <div
                            class="origin-top-right absolute right-0 mt-1 w-64
                            rounded-md shadow-lg text-left"
                        >
                            <div
                                class="rounded-md bg-white dark:bg-gray-700 dark:text-white shadow-xs"
                            >
                                {#each retro.users as usr, index (usr.id)}
                                    {#if usr.active}
                                        <UserCard
                                            user="{usr}"
                                            showBorder="{index !==
                                                retro.users.length - 1}"
                                        />
                                    {/if}
                                {/each}

                                <div class="p-2">
                                    <InviteUser
                                        hostname="{hostname}"
                                        retroId="{retro.id}"
                                    />
                                </div>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
    <div
        class="px-6 py-2 bg-gray-100 dark:bg-gray-800 border-b border-gray-400 dark:border-gray-700 flex flex-wrap"
    >
        <div class="w-1/2">
            <div class="flex items-center text-gray-500 dark:text-gray-300">
                <div
                    class="flex-initial px-1 {retro.phase === 'intro' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Prime Directive
                </div>
                <div class="flex-initial px-1">
                    <ChevronRight />
                </div>
                <div
                    class="flex-initial px-1 {retro.phase === 'brainstorm' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Brainstorm
                </div>
                <div class="flex-initial px-1">
                    <ChevronRight />
                </div>
                <div
                    class="flex-initial px-1 {retro.phase === 'group' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Group
                </div>
                <div class="flex-initial px-1">
                    <ChevronRight />
                </div>
                <div
                    class="flex-initial px-1 {retro.phase === 'vote' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Vote
                </div>
                <div class="flex-initial px-1">
                    <ChevronRight />
                </div>
                <div
                    class="flex-initial px-1 {retro.phase === 'action' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Action Items
                </div>
                <div class="flex-initial px-1">
                    <ChevronRight />
                </div>
                <div
                    class="flex-initial px-1 {retro.phase === 'completed' &&
                        'border-b-2 border-blue-500 dark:border-yellow-400 text-gray-800 dark:text-gray-200'}"
                >
                    Done
                </div>
            </div>
        </div>
        <div class="w-1/2 text-right text-gray-600 dark:text-gray-400">
            {#if retro.phase === 'brainstorm'}
                Add your comments below
            {:else if retro.phase === 'group'}
                Drag and drop comments to group them together
            {:else if retro.phase === 'vote'}
                Vote for the groups you'd like to discuss most
            {:else if retro.phase === 'action'}
                Add action items, you can no longer group or vote comments
            {/if}
        </div>
    </div>
    {#if showExport}
        <Export retro="{retro}" />
    {/if}
    {#if !showExport}
        <div class="w-full p-4 flex flex-wrap">
            {#if retro.phase === 'intro'}
                <div
                    class="m-auto w-full md:w-3/4 lg:w-2/3 md:py-14 lg:py-20 dark:text-white"
                >
                    <h2
                        class="text-3xl md:text-4xl lg:text-5xl font-rajdhani mb-2 tracking-wide"
                    >
                        The Prime Directive
                    </h2>
                    <div class="title-line bg-yellow-thunder"></div>
                    <p
                        class="md:leading-loose tracking-wider text-xl md:text-2xl lg:text-3xl"
                    >
                        "Regardless of what we discover, we understand and truly
                        believe that everyone did the best job they could, given
                        what they knew at the time, their skills and abilities,
                        the resources available, and the situation at hand."
                    </p>
                    <p class="tracking-wider md:text-lg lg:text-xl ">
                        &mdash;Norm Kerth, Project Retrospectives: A Handbook
                        for Team Review <a
                            href="https://retrospectivewiki.org/index.php?title=The_Prime_Directive"
                            target="_blank"
                            class="text-blue-500 hover:text-blue-800 dark:text-sky-400 dark:hover:text-sky-600"
                        >
                            <ExternalLinkIcon class="w-6 h-6 md:w-8 md:h-8" />
                        </a>
                    </p>
                </div>
            {/if}
            {#if retro.phase === 'brainstorm'}
                <div class="w-full grid gap-4 grid-cols-3">
                    <RetroItemForm
                        handleSubmit="{handleItemAdd}"
                        handleDelete="{handleItemDelete}"
                        itemType="worked"
                        newItemPlaceholder="What worked well..."
                        phase="{retro.phase}"
                        isOwner="{isOwner}"
                        items="{workedItems}"
                    />
                    <RetroItemForm
                        handleSubmit="{handleItemAdd}"
                        handleDelete="{handleItemDelete}"
                        itemType="improve"
                        newItemPlaceholder="What needs improvement..."
                        phase="{retro.phase}"
                        isOwner="{isOwner}"
                        items="{improveItems}"
                    />
                    <RetroItemForm
                        handleSubmit="{handleItemAdd}"
                        handleDelete="{handleItemDelete}"
                        itemType="question"
                        newItemPlaceholder="I want to ask..."
                        phase="{retro.phase}"
                        isOwner="{isOwner}"
                        items="{questionItems}"
                    />
                </div>
            {/if}
            {#if retro.phase === 'group'}
                <div
                    class="w-full grid grid-cols-2 md:grid-cols-4 gap-2 md:gap-4"
                >
                    <GroupPhase
                        groups="{groupedItems}"
                        handleItemChange="{handleItemGroupChange}"
                        handleGroupNameChange="{handleGroupNameChange}"
                    />
                </div>
            {/if}
            {#if retro.phase === 'vote'}
                <div class="w-full">
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-2 md:gap-4">
                        <VotePhase
                            groups="{groupedItems}"
                            handleVote="{handleVote}"
                            handleVoteSubtract="{handleVoteSubtract}"
                            voteLimitReached="{voteLimitReached}"
                        />
                    </div>
                </div>
            {/if}
            {#if retro.phase === 'action' || retro.phase === 'completed'}
                <div class="w-full md:w-2/3">
                    <div class="grid grid-cols-2 md:grid-cols-3 gap-2 md:gap-4">
                        <VotePhase groups="{groupedItems}" />
                    </div>
                </div>
                <div class="w-full md:w-1/3">
                    <div class="pl-4">
                        <div class="flex items-center mb-4">
                            <div class="flex-shrink pr-2">
                                <CheckCircle
                                    class="w-8 h-8 text-indigo-500 dark:text-violet-400"
                                />
                            </div>
                            <div class="flex-grow">
                                <form on:submit="{handleActionItem}">
                                    <input
                                        bind:value="{actionItem}"
                                        placeholder="Action item..."
                                        class="dark:bg-gray-800 border-gray-300 dark:border-gray-700 border-2 appearance-none rounded py-2
                    px-3 text-gray-700 dark:text-gray-400 leading-tight focus:outline-none
                    focus:bg-white dark:focus:bg-gray-700 focus:border-indigo-500 dark:focus:border-yellow-400 w-full"
                                        id="actionItem"
                                        name="actionItem"
                                        type="text"
                                        required
                                        disabled="{!isOwner}"
                                    />
                                    <button type="submit" class="hidden"
                                    ></button>
                                </form>
                            </div>
                        </div>
                        {#each retro.actionItems as item, i}
                            <div
                                class="mb-2 p-2 bg-white dark:bg-gray-800 shadow border-l-4 border-indigo-500 dark:border-violet-400"
                            >
                                <div class="flex items-center">
                                    <div class="flex-shrink">
                                        {#if isOwner}
                                            <button
                                                on:click="{handleActionDelete(
                                                    item.id,
                                                )}"
                                                class="pr-2 pt-1 text-gray-500 dark:text-gray-400
                                                hover:text-red-500"
                                            >
                                                <TrashIcon />
                                            </button>
                                        {/if}
                                    </div>
                                    <div class="flex-grow dark:text-white">
                                        <div class="pr-2">
                                            {item.content}
                                        </div>
                                    </div>
                                    <div class="flex-shrink">
                                        <input
                                            type="checkbox"
                                            id="{i}Completed"
                                            checked="{item.completed}"
                                            class="opacity-0 absolute h-6 w-6"
                                            on:change="{handleActionUpdate(
                                                item.id,
                                                item.completed,
                                                item.content,
                                            )}"
                                        />
                                        <div
                                            class="bg-white dark:bg-gray-800 border-2 rounded-md
                                            border-gray-400 dark:border-gray-300 w-6 h-6 flex flex-shrink-0
                                            justify-center items-center mr-2
                                            focus-within:border-blue-500 dark:focus-within:border-sky-500"
                                        >
                                            <CheckboxIcon />
                                        </div>
                                        <label
                                            for="{i}Completed"
                                            class="select-none"></label>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>
    {/if}
{:else}
    <PageLayout>
        <div class="flex items-center">
            <div class="flex-1 text-center">
                {#if JoinPassRequired}
                    <div class="flex justify-center">
                        <div class="w-full md:w-1/2 lg:w-1/3">
                            <form
                                on:submit="{authRetro}"
                                class="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 mb-4"
                                name="authBattle"
                            >
                                <div class="mb-4">
                                    <label
                                        class="block text-gray-700 dark:text-gray-400 font-bold mb-2"
                                        for="battleJoinCode"
                                    >
                                        {$_('passCodeRequired')}
                                    </label>
                                    <input
                                        bind:value="{joinPasscode}"
                                        placeholder="{$_('enterPasscode')}"
                                        class="bg-gray-100 dark:bg-gray-900 border-gray-200 dark:border-gray-800 border-2 appearance-none
                rounded w-full py-2 px-3 text-gray-700 dark:text-gray-300 leading-tight
                focus:outline-none focus:bg-white dark:focus:bg-gray-700 focus:border-indigo-500 focus:caret-indigo-500 dark:focus:border-yellow-400 dark:focus:caret-yellow-400"
                                        id="battleJoinCode"
                                        name="battleJoinCode"
                                        type="password"
                                        required
                                    />
                                </div>

                                <div class="text-right">
                                    <SolidButton type="submit"
                                        >{$_('battleJoin')}</SolidButton
                                    >
                                </div>
                            </form>
                        </div>
                    </div>
                {:else if socketReconnecting}
                    <h1
                        class="text-5xl text-orange-500 leading-tight font-bold"
                    >
                        Ooops, reloading Retro...
                    </h1>
                {:else if socketError}
                    <h1 class="text-5xl text-red-500 leading-tight font-bold">
                        Error joining retro, refresh and try again.
                    </h1>
                {:else}
                    <h1 class="text-5xl text-green-500 leading-tight font-bold">
                        Loading Retro...
                    </h1>
                {/if}
            </div>
        </div>
    </PageLayout>
{/if}

{#if showDeleteRetro}
    <DeleteConfirmation
        toggleDelete="{toggleDeleteRetro}"
        handleDelete="{concedeRetro}"
        confirmText="Are you sure you want to delete this retrospective?"
        confirmBtnText="Delete Retro"
    />
{/if}
