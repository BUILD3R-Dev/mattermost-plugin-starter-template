import React from 'react';
import PropTypes from 'prop-types';
import {connect} from 'react-redux';
import {getTickets} from '../../actions/tickets';
import './ticket_list.css';

class TicketList extends React.PureComponent {
    static propTypes = {
        tickets: PropTypes.array.isRequired,
        getTickets: PropTypes.func.isRequired
    };

    componentDidMount() {
        this.props.getTickets();
    }

    render() {
        return (
            <div className='ticket-list'>
                <h4>Support Tickets</h4>
                {this.props.tickets.map(ticket => (
                    <div key={ticket.id} className='ticket-item'>
                        <span className={`status-indicator ${ticket.status}`} />
                        <div className='ticket-content'>
                            <h5>{ticket.subject}</h5>
                            <p>{ticket.description}</p>
                        </div>
                    </div>
                ))}
            </div>
        );
    }
}

export default connect(state => ({
    tickets: state.entities.tickets.list
}), {getTickets})(TicketList);
