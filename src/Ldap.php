<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine;


class Ldap {

    /**
     * @var Config
     */
    private $Config;

    /**
     * @const string
     */
    const PROTOCOL = 'ldap://';

    /**
     * @const string
     */
    const PROTOCOL_SSL = 'ldaps://';

    /**
     * @var string
     */
    private $address;

    /**
     * @var int
     */
    private $port;

    /**
     * @var bool
     */
    private $useSsl;

    /**
     * @var string
     */
    private $bindDn;

    /**
     * @var string
     */
    private $bindPassword;

    /**
     * @var string
     */
    private $filter;

    /**
     * @var string
     */
    private $baseDn;

    /**
     * @var array
     */
    private $attr;

    public function __construct(Config $Config) {
        $this->Config = $Config;
        $this->address = $this->Config->getLdapServer();
        $this->port = $this->Config->getLdapPort();
        $this->useSsl = $this->Config->isLdapUsingSsl();
        $this->bindDn = $this->Config->getBindDn();
        $this->filter = $this->Config->getLdapFilter();
        $this->baseDn = $this->Config->getBaseDn();
        $this->attr = $this->Config->getLdapAttributes();
        $this->bindPassword = $this->Config->getLdapBindPassword();
    }

    /**
     * @return resource
     */
    public function getConnection() {
        $connection = \ldap_connect($this->buildAddress());
        $this->defaultOptions($connection);
        return $connection;
    }

    /**
     * @param string $username
     * @param string $password
     * @return bool
     */
    public function auth($username, $password) {
        $userDn = $this->search($username);
        if ($userDn) {
            return $this->checkUserCredentials($userDn, $password);
        }

        return false;
    }

    /**
     * @param $username
     * @return bool
     * @throws \Exception
     */
    public function search($username) {
        $connection = $this->getConnection();
        if (@\ldap_bind($connection, $this->bindDn, $this->bindPassword)) {
            $filter = sprintf($this->filter, $username);

            $result = @\ldap_search($connection, $this->baseDn, $filter, $this->attr);
            $entries = @\ldap_get_entries($connection, $result);
            \ldap_unbind($connection);

            if (!isset($entries[0]['dn'])) {
                return false;
            }
            return $entries[0]['dn'];
        }
        $errno = \ldap_errno($connection);
        $error = \ldap_error($connection);
        throw new \Exception(sprintf('[%s] %s while ldap_search operation', $errno, $error));
    }

    /**
     * @param $userDn
     * @param $password
     * @return bool
     * @throws \Exception
     */
    public function checkUserCredentials($userDn, $password) {
        $connection = $this->getConnection();
        if (@\ldap_bind($connection, $userDn, $password)) {
            \ldap_unbind($connection);
            return true;
        }
        $errno = \ldap_errno($connection);
        $error = \ldap_error($connection);
        throw new \Exception(sprintf('[%s] %s while user credentials validation', $errno, $error));
    }

    public function buildAddress() {
        $protocol = self::PROTOCOL;
        if ($this->useSsl) {
            $protocol = self::PROTOCOL_SSL;
        }
        return sprintf('%s%s', $protocol, $this->address);
    }

    /**
     * @param resource $connection
     */
    public function defaultOptions($connection) {
        ldap_set_option($connection, LDAP_OPT_PROTOCOL_VERSION, 3);
        ldap_set_option($connection, LDAP_OPT_REFERRALS, 1);
        ldap_set_option($connection, LDAP_OPT_NETWORK_TIMEOUT, 20);

    }

}
