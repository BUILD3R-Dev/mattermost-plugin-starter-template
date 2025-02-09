// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import type {Store, Action} from 'redux';

import type {GlobalState} from '@mattermost/types/store';

import manifest from '@/manifest';
import type {PluginRegistry} from '@/types/mattermost-webapp';
import TicketingSystemLink from './components/TicketingSystemLink';

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        registry.registerChannelHeaderButtonAction(
            // Use an arrow function to call window.location.href directly
            <TicketingSystemLink onClick={() => {
                window.location.href = '/plugins/com.mattermost.sample.ticketing/';
            }}/>,
            () => {
                // This callback is not used in this case, but it's required by the API
                // We can leave it empty or log a message if needed
            },
            'Tickets'
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
