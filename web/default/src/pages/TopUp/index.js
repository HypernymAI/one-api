import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';

const TopUp = () => {
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [user, setUser] = useState({});

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo('Please enter the top-up code!')
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess('Top-up successful!');
        setUserQuota((quota) => {
          return quota + data;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError('Request failed');
    } finally {
      setIsSubmitting(false); 
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError('Super admin has not set the top-up link!');
      return;
    }
    let url = new URL(topUpLink);
    let username = user.username;
    let user_id = user.id;
    // add  username and user_id to the topup link
    url.searchParams.append('username', username);
    url.searchParams.append('user_id', user_id);
    url.searchParams.append('transaction_id', crypto.randomUUID());
    window.open(url.toString(), '_blank');
  };

  const getUserQuota = async ()=>{
    let res  = await API.get(`/api/user/self`);
    const {success, message, data} = res.data;
    if (success) {
      setUserQuota(data.quota);
      setUser(data);
    } else {
      showError(message);
    }
  }

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    getUserQuota().then();
  }, []);

  return (
    <Segment>
      <Header as='h3'>Top Up Balance</Header>
      <Grid columns={2} stackable>
        <Grid.Column>
          <Form>
            <Form.Input
              placeholder='Redemption Code'
              name='redemptionCode'
              value={redemptionCode}
              onChange={(e) => {
                setRedemptionCode(e.target.value);
              }}
            />
            <Button color='green' onClick={openTopUpLink}>
              Top Up
            </Button>
            <Button color='yellow' onClick={topUp} disabled={isSubmitting}>
                {isSubmitting ? 'Redeeming...' : 'Redeem'}
            </Button>
          </Form>
        </Grid.Column>
        <Grid.Column>
          <Statistic.Group widths='one'>
            <Statistic>
              <Statistic.Value>{renderQuota(userQuota)}</Statistic.Value>
              <Statistic.Label>Remaining Balance</Statistic.Label>
            </Statistic>
          </Statistic.Group>
        </Grid.Column>
      </Grid>
    </Segment>
  );
};

export default TopUp;