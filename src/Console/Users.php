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

namespace Statusengine\Console;


use Statusengine\Backend\StorageBackend;
use Statusengine\Config;
use Statusengine\ValueObjects\User;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Helper\Table;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Question\Question;

class Users extends Command {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var Config
     */
    private $Config;

    public function __construct(Config $Config, StorageBackend $StorageBackend, $name = null) {
        parent::__construct($name);
        $this->StorageBackend = $StorageBackend;
        $this->Config = $Config;
    }

    protected function configure() {

        $this
            // the name of the command (the part after "bin/console")
            ->setName('users')
            // the short description shown while running "php bin/console list"
            ->setDescription('Interface to crate, edit and delete users for Statusengine UI')
            // the full command description shown when running the command with
            // the "--help" option
            ->setHelp("With this CLI tools, you can create, edit and delete users for the Statusengine UI");

        $this->addOption('username', null, InputOption::VALUE_OPTIONAL, 'A username you want to manipulate');
        $this->addOption('password', null, InputOption::VALUE_OPTIONAL, 'A new password for add or edit action');
        $this->addArgument('action', InputArgument::OPTIONAL, 'add, edit, delete or list', 'list');
    }

    public function execute(InputInterface $input, OutputInterface $output) {
        $UserLoader = $this->StorageBackend->getUserLoader();

        if($this->Config->getAuthType() !== 'basic'){
            throw new \RuntimeException(sprintf(
                'Auth type is set to %s. This CLI only address basic auth.',
                $this->Config->getAuthType()
            ));
        }

        switch ($input->getArgument('action')) {
            case 'add':
                $helper = $this->getHelper('question');
                $username = $input->getOption('username');
                if($username === null){
                    $username = '';
                }
                $password = $input->getOption('password');
                if($password === null){
                    $password = '';
                }

                if (mb_strlen($username) == 0) {
                    $question = new Question('Please enter a username' . PHP_EOL);
                    $question->setValidator(function ($answer) {
                        if (!is_string($answer) || mb_strlen($answer) == 0) {
                            throw new \RuntimeException(
                                'Please type in a username'
                            );
                        }
                        return $answer;
                    });

                    $username = $helper->ask($input, $output, $question);
                }

                if (mb_strlen($password) == 0) {
                    $question = new Question('Please enter your password (it is hidden on CLI)' . PHP_EOL);
                    $question->setHidden(true);
                    $question->setHiddenFallback(false);
                    $question->setValidator(function ($answer) {
                        if (!is_string($answer) || mb_strlen($answer) == 0) {
                            throw new \RuntimeException(
                                'Please type in a password'
                            );
                        }
                        return $answer;
                    });
                    $password = $helper->ask($input, $output, $question);
                }

                $User = new User($username, null);
                $User->hashPassword($password);

                $UserLoader->addUser($User->getUsername(), $User->getPassword());

                $output->writeln(sprintf(
                    '<info>User %s created successfully</info>',
                    $User->getUsername()
                ));

                break;

            case 'edit':
                $helper = $this->getHelper('question');
                $username = $input->getOption('username');
                if($username === null){
                    $username = '';
                }
                $password = $input->getOption('password');
                if($password === null){
                    $password = '';
                }

                if (mb_strlen($username) == 0) {
                    $question = new Question('Please enter the username of the user, you want to modify' . PHP_EOL);
                    $question->setAutocompleterValues($this->getUsersForAutocompletion());
                    $question->setValidator(function ($answer) {
                        if (!is_string($answer) || mb_strlen($answer) == 0) {
                            throw new \RuntimeException(
                                'Please type in a username'
                            );
                        }
                        return $answer;
                    });

                    $username = $helper->ask($input, $output, $question);
                }

                if (mb_strlen($password) == 0) {
                    $question = new Question('Please enter the new password (it is hidden on CLI)' . PHP_EOL);
                    $question->setHidden(true);
                    $question->setHiddenFallback(false);
                    $question->setValidator(function ($answer) {
                        if (!is_string($answer) || mb_strlen($answer) == 0) {
                            throw new \RuntimeException(
                                'Please type in a password'
                            );
                        }
                        return $answer;
                    });
                    $password = $helper->ask($input, $output, $question);
                }

                $User = new User($username, null);
                $User->hashPassword($password);

                $UserLoader->changePassword($User->getUsername(), $User->getPassword());

                $output->writeln(sprintf(
                    '<info>User %s updated successfully</info>',
                    $User->getUsername()
                ));

                break;

            case 'delete':
                $helper = $this->getHelper('question');
                $username = $input->getOption('username');
                if($username === null){
                    $username = '';
                }
                if (mb_strlen($username) == 0) {
                    $question = new Question('Please enter a username' . PHP_EOL);
                    $question->setAutocompleterValues($this->getUsersForAutocompletion());
                    $question->setValidator(function ($answer) {
                        if (!is_string($answer) || mb_strlen($answer) == 0) {
                            throw new \RuntimeException(
                                'Please type in a username'
                            );
                        }
                        return $answer;
                    });

                    $username = $helper->ask($input, $output, $question);
                }

                $UserLoader->deleteUser($username);

                $output->writeln(sprintf(
                    '<info>User %s was deleted successfully</info>',
                    $username
                ));
                break;

            default:
                $users = [];
                foreach ($UserLoader->getAllUsers() as $user) {
                    $users[] = [
                        $user['username'],
                        $user['password']
                    ];
                }

                $table = new Table($output);
                $table
                    ->setHeaders(array('Username', 'Password'))
                    ->setRows($users);
                $table->render();
                break;
        }
        return 0;
    }

    /**
     * @return array
     */
    private function getUsersForAutocompletion(){
        $usernames = [];
        $UserLoader = $this->StorageBackend->getUserLoader();
        foreach ($UserLoader->getAllUsers() as $user) {
            $usernames[] = $user['username'];
        }
        return $usernames;
    }

}
