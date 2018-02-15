<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
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


use Statusengine\Exceptions\FileNotFoundException;
use Statusengine\Exceptions\MissingConfigurationItemException;
use Symfony\Component\Yaml\Parser;

class Config {

    /**
     * @var array
     */
    private $config;

    /**
     * @var string
     */
    private $path;

    /**
     * Config constructor.
     * @param null $path
     * @throws FileNotFoundException
     */
    public function __construct($path = null) {
        //default path
        $this->path = __DIR__ . DS . '..' . DS . 'etc' . DS . 'config.yml';

        if ($path !== null) {
            $this->path = $path;
        }


        if (!file_exists($this->path)) {
            throw new FileNotFoundException(sprintf('Config file %s not found or not readable', $this->path));
        }

        $this->parse();
    }

    /**
     * @return void
     */
    public function parse() {
        $yaml = new Parser();
        $config = $yaml->parse(file_get_contents($this->path));

        $this->config = $config;
    }

    /**
     * @return bool
     */
    public function isCrateEnabled() {
        $default = false;
        if (isset($this->config['use_crate'])) {
            return (bool)$this->config['use_crate'];
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function isMysqlEnabled() {
        $default = false;
        if (isset($this->config['use_mysql'])) {
            return (bool)$this->config['use_mysql'];
        }
        return $default;
    }

    /**
     * @return array
     */
    public function getCrateConfig() {
        $default = ['127.0.0.1:4200'];

        if (isset($this->config['crate']['nodes'])) {
            if (is_array($this->config['crate']['nodes']) && !empty($this->config['crate']['nodes'])) {
                return $this->config['crate']['nodes'];
            }
        }

        return $default;
    }

    /**
     * @return array
     */
    public function getMysqlConfig() {
        $config = [
            'host' => '127.0.0.1',
            'port' => 3306,
            'username' => 'statusengine',
            'password' => 'password',
            'database' => 'statusengine_data'
        ];

        foreach ($config as $key => $value) {
            if (isset($this->config['mysql'][$key])) {
                $config[$key] = $this->config['mysql'][$key];
            }
        }

        return $config;
    }

    /**
     * @return bool
     */
    public function isAnonymousAllowed() {
        $default = false;
        if (isset($this->config['allow_anonymous'])) {
            return (bool)$this->config['allow_anonymous'];
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function canAnonymousSubmitCommand() {
        $default = false;
        if (isset($this->config['anonymous_can_submit_commands'])) {
            return (bool)$this->config['anonymous_can_submit_commands'];
        }
        return $default;
    }

    /**
     * @return array
     */
    public function getUrlsWithoutLogin() {
        $default = ['/login'];
        if (isset($this->config['urls_without_login']) && is_array($this->config['urls_without_login'])) {
            return $this->config['urls_without_login'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getAuthType() {
        $default = 'basic';
        $possibleValues = ['basic', 'ldap'];
        if (isset($this->config['auth_type'])) {
            $authType = mb_strtolower($this->config['auth_type']);
            if (in_array($authType, $possibleValues, true)) {
                return $authType;
            }
        }

        return $default;
    }

    /**
     * @return string
     * @throws MissingConfigurationItemException
     */
    public function getLdapServer() {
        if (isset($this->config['ldap_server'])) {
            return (string)$this->config['ldap_server'];
        }

        throw new MissingConfigurationItemException('Key ldap_server not found in configuration file');
    }

    /**
     * @return int
     */
    public function getLdapPort() {
        $default = 389;
        if (isset($this->config['ldap_port'])) {
            if (is_numeric($this->config['ldap_port'])) {
                return (int)$this->config['ldap_port'];
            }
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function isLdapUsingSsl() {
        $default = false;
        if (isset($this->config['ldap_use_ssl'])) {
            return (bool)$this->config['ldap_use_ssl'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getBindDn() {
        $default = 'cn=ldapsearch,dc=example,dc=com';
        if (isset($this->config['ldap_bind_dn'])) {
            return (string)$this->config['ldap_bind_dn'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getLdapBindPassword() {
        $default = 'password';
        if (isset($this->config['ldap_bind_password'])) {
            return (string)$this->config['ldap_bind_password'];
        }
        return $default;
    }

    public function getBaseDn() {
        $default = 'dc=example,dc=com';
        if (isset($this->config['ldap_base_dn'])) {
            return (string)$this->config['ldap_base_dn'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getLdapFilter() {
        $default = '(sAMAccountName=%s)';
        if (isset($this->config['ldap_filter'])) {
            return (string)$this->config['ldap_filter'];
        }
        return $default;
    }

    /**
     * @return array
     */
    public function getLdapAttributes() {
        $default = ['memberof'];
        if (isset($this->config['ldap_attribute'])) {
            return (array)$this->config['ldap_attribute'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getGraphiteUrl() {
        $default = "http://localhost:8080";
        if (isset($this->config['graphite_url'])) {
            return (string)$this->config['graphite_url'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getGraphiteIllegalCharacters() {
        $default = "/[^a-zA-Z^0-9\-\.]/";
        if (isset($this->config['graphite_illegal_characters'])) {
            return (string)$this->config['graphite_illegal_characters'];
        }
        return $default;
    }


    /**
     * @return string
     */
    public function getGraphitePrefix() {
        $default = "statusengine";
        if (isset($this->config['graphite_prefix'])) {
            return (string)$this->config['graphite_prefix'];
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function getGraphiteUseBasicAuth() {
        $default = false;
        if (isset($this->config['graphite_use_basic_auth'])) {
            return (bool)$this->config['graphite_use_basic_auth'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getGraphiteUser() {
        $default = "graphite";
        if (isset($this->config['graphite_user'])) {
            return (string)$this->config['graphite_user'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getGraphitePassword() {
        $default = "password";
        if (isset($this->config['graphite_password'])) {
            return (string)$this->config['graphite_password'];
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function getGraphiteAllowSelfSignedCertificates() {
        $default = false;
        if (isset($this->config['graphite_allow_self_signed_certificates'])) {
            return (bool)$this->config['graphite_allow_self_signed_certificates'];
        }
        return $default;
    }

    /**
     * @return bool
     */
    public function getDisplayPerfdata() {
        $default = false;
        if (isset($this->config['display_perfdata'])) {
            return (bool)$this->config['display_perfdata'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getElasticsearchIndex() {
        $default = 'statusengine-metric';
        if (isset($this->config['elasticsearch_index'])) {
            return (string)$this->config['elasticsearch_index'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getElasticsearchAddress() {
        $default = '127.0.0.1';
        if (isset($this->config['elasticsearch_address'])) {
            return (string)$this->config['elasticsearch_address'];
        }
        return $default;
    }

    /**
     * @return int
     */
    public function getElasticsearchPort() {
        $default = 9200;
        if (isset($this->config['elasticsearch_port'])) {
            return (int)$this->config['elasticsearch_port'];
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getElasticsearchPattern() {
        $default = 'none';
        $patterns = [
            'none',
            'daily',
            'weekly',
            'monthly'
        ];
        if (isset($this->config['elasticsearch_pattern'])) {
            if (in_array($this->config['elasticsearch_pattern'], $patterns, true)) {
                return $this->config['elasticsearch_pattern'];
            }
        }
        return $default;
    }

    /**
     * @return string
     */
    public function getPerfdataBackend() {
        $default = 'crate';
        $availableBackend = ['crate', 'graphite', 'mysql', 'elasticsearch'];
        if (isset($this->config['perfdata_backend'])) {
            $value = (string)$this->config['perfdata_backend'];
            if (in_array($value, $availableBackend, true)) {
                return $value;
            }
        }
        return $default;
    }

}
