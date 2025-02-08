import TicketList from './components/ticket_list/ticket_list';
import {openTicketModal} from './actions/modals';

export default class SupportTicketPlugin {
    initialize(registry, store) {
        // Register the ticket list in the channel sidebar
        registry.registerRootComponent(TicketList);
        
        // Add ticket creation button to channel header
        registry.registerChannelHeaderButtonAction(
            <span className='icon fa fa-ticket'/>,
            () => store.dispatch(openTicketModal()),
            'Create Ticket'
        );
    }
}
