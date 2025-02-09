import React from 'react';

interface Props {
    onClick: () => void;
}
const TicketingSystemLink: React.FC<Props> = ({onClick}) => {

    return (
        <button onClick={onClick}  className="channel-header__trigger style--none">
            <i className="icon fa fa-ticket"/>
            <span className="channel-header__title">Tickets</span>
        </button>
    )
}

export default TicketingSystemLink;
