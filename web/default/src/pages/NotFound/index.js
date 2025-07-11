import React from 'react';
import { Message } from 'semantic-ui-react';

const NotFound = () => (
  <>
    <Message negative>
      <Message.Header>Page Not Found</Message.Header>
      <p>Please check if your browser address is correct</p>
    </Message>
  </>
);

export default NotFound;
